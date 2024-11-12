package routes

import (
	"github.com/gofiber/fiber/v2"
	api "kokal5296/web/handlers"
)

const (
	userPath       = "/user"
	bookPath       = "/book"
	bookBorrowPath = "/book_borrow"
)

// SetupRoutes initializes all routes for the application
func SetupRoutes(app *fiber.App, userHandler api.UserApi, bookHandler api.BookApi, bookBorrowHandler api.BookBorrowApi) {
	setupUserRoutes(app, userHandler)
	setupBookRoutes(app, bookHandler)
	setupBookBorrowRoutes(app, bookBorrowHandler)
}

func setupUserRoutes(app *fiber.App, handler api.UserApi) {
	app.Post(userPath, handler.CreateUser)
	app.Get(userPath+"/:id", handler.GetUser)
	app.Get(userPath+"s", handler.GetAllUsers)
	app.Put(userPath+"/:id", handler.UpdateUser)
	app.Delete(userPath+"/:id", handler.DeleteUser)
}

func setupBookRoutes(app *fiber.App, handler api.BookApi) {
	app.Post(bookPath, handler.CreateBook)
	app.Get(bookPath+"/:id", handler.GetBook)
	app.Get(bookPath+"s", handler.GetAllBooks)
	app.Put(bookPath+"/:id", handler.UpdateBook)
	app.Delete(bookPath+"/:id", handler.DeleteBook)
}

func setupBookBorrowRoutes(app *fiber.App, handler api.BookBorrowApi) {
	app.Get(bookBorrowPath, handler.GetAvailableBooks)
	app.Get(bookBorrowPath+"ed", handler.AllBorrowedBooks)
	app.Post(bookBorrowPath, handler.BorrowBook)
	app.Put(bookBorrowPath, handler.ReturnBook)
}
