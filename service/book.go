package service

import (
	"context"
	"fmt"
	"kokal5296/database"
	er "kokal5296/errors"
	"kokal5296/models/book"
	"log"
	"time"
)

type BookServiceStruct struct {
	dbService database.DatabaseService
}

const bookService = "bookService - "

// BookService interface defines methods for book-related operations
type BookService interface {
	CreateBook(ctx context.Context, newBook book.Book) error
	GetBook(ctx context.Context, bookId int) (*book.Book, error)
	GetAllBooks(ctx context.Context) ([]book.Book, error)
	UpdateBook(ctx context.Context, bookId int, updatedBook book.Book) error
	DeleteBook(ctx context.Context, bookId int) error
}

// NewBookService creates a new instance of BookServiceStruct, implementing BookService
func NewBookService(dbService database.DatabaseService) BookService {
	return &BookServiceStruct{
		dbService: dbService,
	}
}

// CreateBook creates a new book in the database
func (s *BookServiceStruct) CreateBook(ctx context.Context, newBook book.Book) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "CreateBook"

	log.Printf("book: %v", newBook)
	err := s.titleExists(ctx, newBook)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	query := `INSERT INTO books (title, quantity) VALUES ($1, $2)`
	_, err = s.dbService.GetPool().Exec(ctx, query, newBook.Title, newBook.Quantity)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error creating book: %v", err)
		return er.Wrap(funcName, err)
	}

	return nil
}

// GetBook retrieves a book from the database by its ID
func (s *BookServiceStruct) GetBook(ctx context.Context, bookId int) (*book.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "GetBook"

	var book book.Book
	query := `SELECT id, title, quantity FROM books WHERE id = $1`
	err := s.dbService.GetPool().QueryRow(ctx, query, bookId).Scan(&book.ID, &book.Title, &book.Quantity)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		log.Printf("Error getting book: %v", err)
		return nil, er.Wrap(funcName, err)
	}

	return &book, nil
}

// GetAllBooks retrieves all books from the database
func (s *BookServiceStruct) GetAllBooks(ctx context.Context) ([]book.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "GetAllBooks"

	var books []book.Book
	query := `SELECT id, title, quantity FROM books`
	rows, err := s.dbService.GetPool().Query(ctx, query)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		log.Printf("Error getting books: %v", err)
		return nil, er.Wrap(funcName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var book book.Book
		err := rows.Scan(&book.ID, &book.Title, &book.Quantity)
		if err != nil {
			log.Printf("Error scanning books: %v", err)
			return nil, er.Wrap(funcName, err)
		}
		books = append(books, book)
	}

	return books, nil
}

// UpdateBook updates a book in the database by its ID.
// It checks if the book exists and if the title and id of the book match, and if the do match,
// it doesn't check if the title already exists in the database, because that means we are updating same book.
// Which allows to change only quantity of the book.
// If the title is different, it checks if the title already exists.
func (s *BookServiceStruct) UpdateBook(ctx context.Context, bookId int, updatedBook book.Book) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "UpdateBook"

	err := s.bookExists(ctx, bookId)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	ok, err := s.bookAndTitleMatch(ctx, bookId, updatedBook)
	if err != nil {
		return er.Wrap(funcName, err)
	}
	if !ok {
		err = s.titleExists(ctx, updatedBook)
		if err != nil {
			return er.Wrap(funcName, err)
		}
	}

	query := `UPDATE books SET title = $1, quantity = $2 WHERE id = $3`
	_, err = s.dbService.GetPool().Exec(ctx, query, updatedBook.Title, updatedBook.Quantity, bookId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error updating book: %v", err)
		return er.Wrap(funcName, err)
	}

	return nil
}

// DeleteBook deletes a book from the database by its ID, if it exists
func (s *BookServiceStruct) DeleteBook(ctx context.Context, bookId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "DeleteBook"

	query := `DELETE FROM books WHERE id = $1`
	_, err := s.dbService.GetPool().Exec(ctx, query, bookId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error deleting book: %v", err)
		return er.Wrap(funcName, err)
	}

	return nil
}

// bookExists checks if a book with the given ID exists in the database
func (s *BookServiceStruct) bookExists(ctx context.Context, bookId int) error {
	funcName := bookService + "bookExists"

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM books WHERE id = $1)`
	err := s.dbService.GetPool().QueryRow(ctx, query, bookId).Scan(&exists)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error checking if book exists: %v", err)
		return er.Wrap(funcName, err)
	}

	if !exists {
		message := fmt.Sprintf("Book with id %d does not exist", bookId)
		return er.New(funcName, message, nil)
	}

	return nil
}

// titleExists checks if a book with the given title exists in the database
func (s *BookServiceStruct) titleExists(ctx context.Context, book book.Book) error {
	funcName := bookService + "titleExists"

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM books WHERE title = $1)`
	err := s.dbService.GetPool().QueryRow(ctx, query, book.Title).Scan(&exists)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error checking if title of the book exists: %v", err)
		return er.Wrap(funcName, err)
	}

	log.Printf("exists: %v", exists)
	if exists {
		message := fmt.Sprintf("Book with title %s, already exists", book.Title)
		return er.New(funcName, message, nil)
	}

	return nil
}

// bookAndTitleMatch checks if the book with the given ID and title match in the database
func (s *BookServiceStruct) bookAndTitleMatch(ctx context.Context, bookId int, book book.Book) (bool, error) {
	funcName := bookService + "bookAndTitleExists"

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM books WHERE id = $1 AND title = $2)`
	err := s.dbService.GetPool().QueryRow(ctx, query, bookId, book.Title).Scan(&exists)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return false, er.Wrap(funcName, err)
		}
		log.Printf("Error checking if book and title match: %v", err)
		return false, er.Wrap(funcName, err)
	}

	if !exists {
		return false, nil
	}

	return true, nil
}
