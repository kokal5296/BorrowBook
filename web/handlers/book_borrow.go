package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	er "kokal5296/errors"
	"kokal5296/models/book_borrow"
	"kokal5296/service"
	validate "kokal5296/web/validation"
	"log"
	"net/http"
)

type BookBorrowApiStruct struct {
	bookBorrowService service.BookBorrowService
}

// NewBookBorrowApiService creates a new instance of BookBorrowApiStruct, which implements the BookBorrowApi interface
func NewBookBorrowApiService(bookBorrowService service.BookBorrowService) BookBorrowApi {
	return &BookBorrowApiStruct{
		bookBorrowService: bookBorrowService,
	}
}

// GetAvailableBooks handles the request to get all available books
func (s *BookBorrowApiStruct) GetAvailableBooks(c *fiber.Ctx) error {

	log.Println("Requesting to get available books")

	funcName := handler + "GetAvailableBooks"

	books, err := s.bookBorrowService.GetAvailableBooks(c.Context())
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).JSON(books)
}

// AllBorrowedBooks handles the request to get all borrowed books
func (s *BookBorrowApiStruct) AllBorrowedBooks(c *fiber.Ctx) error {

	log.Println("Requesting to get all borrowed books")

	funcName := handler + "AllBorrowedBooks"

	books, err := s.bookBorrowService.AllBorrowedBooks(c.Context())
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).JSON(books)
}

// BorrowBook handles the request to borrow a book
func (s *BookBorrowApiStruct) BorrowBook(c *fiber.Ctx) error {

	log.Println("Requesting to borrow book")
	var bookBorrow book_borrow.BookBorrow

	funcName := handler + "BorrowBook"

	err := json.Unmarshal(c.Body(), &bookBorrow)
	if err != nil {
		log.Printf("Error while unmarshalling book borrow: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateBookBorrow(bookBorrow)
	if validateErr != nil {
		log.Printf("Error while validating book borrow: %v", validateErr)
		return c.Status(fiber.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.bookBorrowService.BorrowBook(c.Context(), bookBorrow.BookID, bookBorrow.UserID)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).SendString("Book was successfully borrowed")
}

// ReturnBook handles the request to return a book
func (s *BookBorrowApiStruct) ReturnBook(c *fiber.Ctx) error {

	log.Println("Requesting to return book")
	var bookBorrow book_borrow.BookBorrow

	funcName := handler + "ReturnBook"

	err := json.Unmarshal(c.Body(), &bookBorrow)
	if err != nil {
		log.Printf("Error while unmarshalling book borrow: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateBookBorrow(bookBorrow)
	if validateErr != nil {
		log.Printf("Error while validating book borrow: %v", validateErr)
		return c.Status(fiber.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.bookBorrowService.ReturnBook(c.Context(), bookBorrow.BookID, bookBorrow.UserID)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(fiber.StatusOK).SendString("Book was successfully returned")
}
