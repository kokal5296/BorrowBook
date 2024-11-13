package service

import (
	"context"
	"github.com/jackc/pgx/v4"
	"kokal5296/database"
	er "kokal5296/errors"
	"kokal5296/models/book"
	"kokal5296/models/book_borrow"
	"log"
	"time"
)

type BookBorrowStruct struct {
	dbService   database.DatabaseService
	BookService BookService
	userService UserService
}

const bookBorrowService = "bookBorrowService - "

// BookBorrowService interface defgines methods for book borrow-related operations
type BookBorrowService interface {
	GetAvailableBooks(ctx context.Context) ([]book.Book, error)
	AllBorrowedBooks(ctx context.Context) ([]book_borrow.BookBorrow, error)
	BorrowBook(ctx context.Context, bookId int, userId int) error
	ReturnBook(ctx context.Context, bookId int, userId int) error
}

// NewBookBorrowService creates a new instance of BookBorrowService, implementing the BookBorrowStruct
func NewBookBorrowService(dbService database.DatabaseService, bookService BookService, userService UserService) BookBorrowService {
	return &BookBorrowStruct{
		dbService:   dbService,
		BookService: bookService,
		userService: userService,
	}
}

// GetAvailableBooks returns all books that are available for borrowing
func (s *BookBorrowStruct) GetAvailableBooks(ctx context.Context) ([]book.Book, error) {
	ctx, cancle := context.WithTimeout(ctx, 5*time.Second)
	defer cancle()

	funcName := bookBorrowService + "GetAvailableBooks"

	var books []book.Book
	query := `SELECT * FROM books WHERE quantity > 0`
	rows, err := s.dbService.GetPool().Query(ctx, query)
	if err != nil {
		if er.HandleDeadlineExceededError(bookBorrowService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		log.Printf("Error getting available books: %v", err)
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

// AllBorrowedBooks returns all books that are currently borrowed and not yet returned
func (s *BookBorrowStruct) AllBorrowedBooks(ctx context.Context) ([]book_borrow.BookBorrow, error) {
	ctx, cancle := context.WithTimeout(ctx, 5*time.Second)
	defer cancle()

	funcName := bookBorrowService + "AllBorrowedBooks"

	var result []book_borrow.BookBorrow
	query := `SELECT id, book_id, user_id, borrow_date, return_date FROM book_borrows WHERE return_date IS NULL`
	rows, err := s.dbService.GetPool().Query(ctx, query)
	if err != nil {
		if er.HandleDeadlineExceededError(bookBorrowService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		log.Printf("Error getting borrowed books: %v", err)
		return nil, er.Wrap(funcName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var bookBorrowed book_borrow.BookBorrow
		err := rows.Scan(&bookBorrowed.ID, &bookBorrowed.BookID, &bookBorrowed.UserID, &bookBorrowed.Borrow_date, &bookBorrowed.Return_date)
		if err != nil {
			log.Printf("Error scanning books: %v", err)
			return nil, er.Wrap(funcName, err)
		}
		result = append(result, bookBorrowed)
	}

	log.Printf("All borrowed books: %v", result)
	return result, nil
}

// BorrowBook allows a user to borrow a book if it's available and the user has not already borrowed it
func (s *BookBorrowStruct) BorrowBook(ctx context.Context, bookId int, userId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := bookService + "BorrowBook"

	var quantity int
	query := `SELECT quantity FROM books WHERE id = $1`
	err := s.dbService.GetPool().QueryRow(ctx, query, bookId).Scan(&quantity)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error getting book: %v", err)
		return er.Wrap(funcName, err)
	}

	if quantity <= 0 {
		message := "Book is not available"
		return er.New(funcName, message, nil)
	}

	err = s.userService.UserExist(ctx, userId)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	var bookBorrowed bool
	query = `SELECT * FROM book_borrows WHERE book_id = $1 AND user_id = $2 AND borrow_date IS NOT NULL AND return_date IS NULL`
	err = s.dbService.GetPool().QueryRow(ctx, query, bookId, userId).Scan(&bookBorrowed)
	if err != nil {
		if err == pgx.ErrNoRows {
			bookBorrowed = false
		} else {
			if er.HandleDeadlineExceededError(bookService, err) != nil {
				return er.Wrap(funcName, err)
			}
			log.Printf("Error getting borrowed book: %v", err)
			return er.Wrap(funcName, err)
		}
	}

	if bookBorrowed {
		message := "Book is already borrowed"
		return er.New(funcName, message, nil)
	}

	query = `INSERT INTO book_borrows (book_id, user_id) VALUES ($1, $2)`
	_, err = s.dbService.GetPool().Exec(ctx, query, bookId, userId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error borrowing book: %v", err)
		return er.Wrap(funcName, err)
	}

	query = `UPDATE books SET quantity = quantity - 1 WHERE id = $1`
	_, err = s.dbService.GetPool().Exec(ctx, query, bookId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error updating book: %v", err)
		return er.Wrap(funcName, err)
	}

	return nil
}

// ReturnBook allows a user to return a book if they have borrowed it
func (s *BookBorrowStruct) ReturnBook(ctx context.Context, bookId int, userId int) error {
	ctx, cancle := context.WithTimeout(ctx, 5*time.Second)
	defer cancle()

	funcName := bookService + "ReturnBook"

	var activeBorrowCount int
	query := `SELECT 1 FROM book_borrows WHERE book_id = $1 AND user_id = $2 AND borrow_date IS NOT NULL AND return_date IS NULL`
	err := s.dbService.GetPool().QueryRow(ctx, query, bookId, userId).Scan(&activeBorrowCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			message := "Book is not borrowed"
			return er.New(funcName, message, nil)
		} else {
			if er.HandleDeadlineExceededError(bookService, err) != nil {
				return er.Wrap(funcName, err)
			}
			log.Printf("Error getting borrowed book: %v", err)
			return er.Wrap(funcName, err)
		}
	}

	if activeBorrowCount == 0 {
		message := "Book is not currently borrowed by the user"
		return er.New(funcName, message, nil)
	}

	query = `UPDATE book_borrows SET return_date = NOW() WHERE book_id = $1 AND user_id = $2 AND return_date IS NULL`
	_, err = s.dbService.GetPool().Exec(ctx, query, bookId, userId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error updating returning book: %v", err)
		return er.Wrap(funcName, err)
	}

	query = `UPDATE books SET quantity = quantity + 1 WHERE id = $1`
	_, err = s.dbService.GetPool().Exec(ctx, query, bookId)
	if err != nil {
		if er.HandleDeadlineExceededError(bookService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error updating book: %v, quantity", err)
		return er.Wrap(funcName, err)
	}

	return nil
}
