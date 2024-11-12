package validate

import (
	"github.com/go-playground/validator/v10"
	"kokal5296/models/book"
	"kokal5296/models/book_borrow"
	"kokal5296/models/user"
)

// Variable with functuion to create new validation
var validate = validator.New()

// validateStruct validates any given struct based on tags defined within the struct
func validateStruct(input interface{}) error {
	return validate.Struct(input)
}

func ValidateUser(user user.User) error {
	return validateStruct(user)
}

func ValidateBook(book book.Book) error {
	return validateStruct(book)
}

func ValidateBookBorrow(book book_borrow.BookBorrow) error {
	return validateStruct(book)
}
