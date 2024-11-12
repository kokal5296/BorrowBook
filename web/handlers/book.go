package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	er "kokal5296/errors"
	"kokal5296/models/book"
	"kokal5296/service"
	validate "kokal5296/web/validation"
	"log"
	"strconv"
)

type BookApiStruct struct {
	bookService service.BookService
}

// NewBookApiService creates a new instance of BookApiStruct, which implements the BookApi interface
func NewBookApiService(bookService service.BookService) BookApi {
	return &BookApiStruct{
		bookService: bookService,
	}
}

// CreateBook handles the request to create a new book
func (s *BookApiStruct) CreateBook(c *fiber.Ctx) error {

	log.Println("Requesting to create new book")
	var newBook book.Book

	funcName := handler + "CreateBook"

	err := json.Unmarshal(c.Body(), &newBook)
	if err != nil {
		log.Printf("Error while unmarshalling book: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateBook(newBook)
	if validateErr != nil {
		log.Printf("Error while validating book: %v", validateErr)
		return c.Status(fiber.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.bookService.CreateBook(c.Context(), newBook)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusCreated).SendString("Book was successfully created")
}

// GetBook handles the request to get a book by id
func (s *BookApiStruct) GetBook(c *fiber.Ctx) error {

	log.Println("Requesting to get book by id")
	funcName := handler + "GetBook"
	id := c.Params("id")

	bookId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Error while converting id to int: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	book, err := s.bookService.GetBook(c.Context(), bookId)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).JSON(book)
}

// GetAllBooks handles the request to get all books
func (s *BookApiStruct) GetAllBooks(c *fiber.Ctx) error {

	log.Println("Requesting to get all books")
	funcName := handler + "GetAllBooks"

	books, err := s.bookService.GetAllBooks(c.Context())
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).JSON(books)
}

// UpdateBook handles the request to update a book
func (s *BookApiStruct) UpdateBook(c *fiber.Ctx) error {

	log.Println("Requesting to update book")
	funcName := handler + "UpdateBook"

	var updateBook book.Book

	id := c.Params("id")

	bookId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = json.Unmarshal(c.Body(), &updateBook)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateBook(updateBook)
	if validateErr != nil {
		return c.Status(fiber.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.bookService.UpdateBook(c.Context(), bookId, updateBook)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).SendString("Book was updated successfully")
}

// DeleteBook handles the request to delete a book
func (s *BookApiStruct) DeleteBook(c *fiber.Ctx) error {

	log.Println("Requesting to delete book")
	funcName := handler + "DeleteBook"

	id := c.Params("id")

	bookId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	err = s.bookService.DeleteBook(c.Context(), bookId)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).SendString("Book was deleted successfully")
}
