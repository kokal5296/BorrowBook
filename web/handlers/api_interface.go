package api

import "github.com/gofiber/fiber/v2"

const handler string = "handler - "

// UserApi defines the interface for handling user related HTTP requests
type UserApi interface {
	CreateUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

// BookApi defines the interface for handling book related HTTP requests
type BookApi interface {
	CreateBook(c *fiber.Ctx) error
	GetBook(c *fiber.Ctx) error
	GetAllBooks(c *fiber.Ctx) error
	UpdateBook(c *fiber.Ctx) error
	DeleteBook(c *fiber.Ctx) error
}

// BookBorrowApi defines the interface for handling book borrow related HTTP requests
type BookBorrowApi interface {
	GetAvailableBooks(c *fiber.Ctx) error
	AllBorrowedBooks(c *fiber.Ctx) error
	BorrowBook(c *fiber.Ctx) error
	ReturnBook(c *fiber.Ctx) error
}
