package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"kokal5296/models/book"
	"kokal5296/models/book_borrow"
	"kokal5296/models/user"
	"kokal5296/service"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAvailibleBooks tests the scenarios for retrieving all available books
func TestAvailibleBooks(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	bookService := service.NewBookService(dbService)
	bookBorrowService := service.NewBookBorrowService(dbService, bookService, userService)
	bookBorrowApi := NewBookBorrowApiService(bookBorrowService)

	app := fiber.New()
	app.Get("/book_borrow", bookBorrowApi.GetAvailableBooks)

	t.Run("Retrieve all available books", func(t *testing.T) {
		existingBooks := []book.Book{
			{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 0},
			{Title: "Lord of the Rings: Two Towers", Quantity: 3},
			{Title: "Lord of the Rings: Return of the King", Quantity: 10},
		}

		for _, b := range existingBooks {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
			assert.NoError(t, err)
		}

		req, _ := http.NewRequest("GET", "/book_borrow", nil)
		resp, _ := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var borrowedBooks []book.Book
		err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
		assert.NoError(t, err)
		assert.Len(t, borrowedBooks, 2)
		log.Printf("Borrowed books: %v", borrowedBooks)
	})

	t.Run("No available books", func(t *testing.T) {
		_, err = dbService.GetPool().Exec(context.Background(), "DELETE FROM books")
		assert.NoError(t, err)

		existingBooks := []book.Book{
			{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 0},
			{Title: "Lord of the Rings: Two Towers", Quantity: 0},
			{Title: "Lord of the Rings: Return of the King", Quantity: 0},
		}

		for _, b := range existingBooks {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
			assert.NoError(t, err)
		}

		req, _ := http.NewRequest("GET", "/book_borrow", nil)
		resp, _ := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var borrowedBooks []book.Book
		err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
		assert.NoError(t, err)
		assert.Len(t, borrowedBooks, 0)
		log.Printf("Borrowed books: %v", borrowedBooks)
	})
}

// TestBorrowBook tests the scenarios for borrowing a book
func TestBorrowBook(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	bookService := service.NewBookService(dbService)
	bookBorrowService := service.NewBookBorrowService(dbService, bookService, userService)
	bookBorrowApi := NewBookBorrowApiService(bookBorrowService)

	app := fiber.New()
	app.Post("/book_borrow", bookBorrowApi.BorrowBook)

	existingBooks := []book.Book{
		{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 5},
		{Title: "Lord of the Rings: Two Towers", Quantity: 0},
		{Title: "Lord of the Rings: Return of the King", Quantity: 10},
	}

	existingUsers := []user.User{
		{FirstName: "Tine", LastName: "Kokalj"},
		{FirstName: "Žan", LastName: "Horvat"},
		{FirstName: "Luka", LastName: "Potočnik"},
	}

	for _, b := range existingBooks {
		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
		assert.NoError(t, err)
	}

	for _, u := range existingUsers {
		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2)", u.FirstName, u.LastName)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		input         book_borrow.BookBorrow
		expected      int
		expectedCount int
	}{
		{
			name:          "Borrow a book",
			input:         book_borrow.BookBorrow{BookID: 1, UserID: 1},
			expected:      http.StatusOK,
			expectedCount: 1,
		},
		{
			name:          "Borrow a book that is not available",
			input:         book_borrow.BookBorrow{BookID: 2, UserID: 1},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Borrow a book that is already borrowed",
			input:         book_borrow.BookBorrow{BookID: 1, UserID: 1},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Borrow a book that does not exist",
			input:         book_borrow.BookBorrow{BookID: 100, UserID: 1},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Borrow a book with a user that does not exist",
			input:         book_borrow.BookBorrow{BookID: 1, UserID: 100},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/book_borrow", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, resp.StatusCode)

			var borrowedBooks int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM book_borrows").Scan(&borrowedBooks)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, borrowedBooks)
		})
	}

}

// TestReturnBook tests the scenarios for returning a book
func TestReturnBook(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	bookService := service.NewBookService(dbService)
	bookBorrowService := service.NewBookBorrowService(dbService, bookService, userService)
	bookBorrowApi := NewBookBorrowApiService(bookBorrowService)

	app := fiber.New()
	app.Put("/book_borrow", bookBorrowApi.ReturnBook)

	existingBooks := []book.Book{
		{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 5},
		{Title: "Lord of the Rings: Two Towers", Quantity: 0},
		{Title: "Lord of the Rings: Return of the King", Quantity: 10},
	}

	existingUsers := []user.User{
		{FirstName: "Tine", LastName: "Kokalj"},
		{FirstName: "Žan", LastName: "Horvat"},
		{FirstName: "Luka", LastName: "Potočnik"},
	}

	for _, b := range existingBooks {
		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
		assert.NoError(t, err)
	}

	for _, u := range existingUsers {
		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2)", u.FirstName, u.LastName)
		assert.NoError(t, err)
	}

	_, err = dbService.GetPool().Exec(context.Background(), "INSERT INTO book_borrows (book_id, user_id) VALUES (1, 1)")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		input         book_borrow.BookBorrow
		expected      int
		expectedCount int
	}{
		{
			name:          "Return a book that is not borrowed",
			input:         book_borrow.BookBorrow{BookID: 3, UserID: 1},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Return a book that does not exist",
			input:         book_borrow.BookBorrow{BookID: 100, UserID: 1},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Return a book with a user that does not exist",
			input:         book_borrow.BookBorrow{BookID: 1, UserID: 100},
			expected:      http.StatusInternalServerError,
			expectedCount: 1,
		},
		{
			name:          "Return a book",
			input:         book_borrow.BookBorrow{BookID: 1, UserID: 1},
			expected:      http.StatusOK,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.input)
			req := httptest.NewRequest("PUT", "/book_borrow", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, resp.StatusCode)

			var borrowedBooks int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM book_borrows WHERE return_date IS NULL").Scan(&borrowedBooks)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, borrowedBooks)
		})
	}
}

// TestAllBorrowedBooks tests the scenarios for retrieving all borrowed books
func TestAllBorrowedBooks(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	bookService := service.NewBookService(dbService)
	bookBorrowService := service.NewBookBorrowService(dbService, bookService, userService)
	bookBorrowApi := NewBookBorrowApiService(bookBorrowService)

	app := fiber.New()
	app.Get("/book_borrowed", bookBorrowApi.AllBorrowedBooks)
	app.Put("/book_borrow", bookBorrowApi.ReturnBook)

	t.Run("Retrieve all borrowed books", func(t *testing.T) {
		existingBooks := []book.Book{
			{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 5},
			{Title: "Lord of the Rings: Two Towers", Quantity: 0},
			{Title: "Lord of the Rings: Return of the King", Quantity: 10},
		}

		existingUsers := []user.User{
			{FirstName: "Tine", LastName: "Kokalj"},
			{FirstName: "Žan", LastName: "Horvat"},
			{FirstName: "Luka", LastName: "Potočnik"},
		}

		for _, b := range existingBooks {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
			assert.NoError(t, err)
		}

		for _, u := range existingUsers {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2)", u.FirstName, u.LastName)
			assert.NoError(t, err)
		}

		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO book_borrows (book_id, user_id) VALUES (1, 1)")
		assert.NoError(t, err)
		_, err = dbService.GetPool().Exec(context.Background(), "INSERT INTO book_borrows (book_id, user_id) VALUES (2, 1)")
		assert.NoError(t, err)

		req, _ := http.NewRequest("GET", "/book_borrowed", nil)
		resp, _ := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var borrowedBooks []book_borrow.BookBorrow
		err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
		assert.NoError(t, err)
		assert.Len(t, borrowedBooks, 2)
	})

	t.Run("No borrowed books", func(t *testing.T) {
		_, err = dbService.GetPool().Exec(context.Background(), "DELETE FROM book_borrows")
		assert.NoError(t, err)

		req, _ := http.NewRequest("GET", "/book_borrowed", nil)
		resp, _ := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var borrowedBooks []book_borrow.BookBorrow
		err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
		assert.NoError(t, err)
		assert.Len(t, borrowedBooks, 0)
	})

	t.Run("Retrieve all borrowed books with some retured", func(t *testing.T) {
		existingBooks := []book.Book{
			{Title: "Lord of the Rings: Fellowship of the Ring", Quantity: 5},
			{Title: "Lord of the Rings: Two Towers", Quantity: 0},
			{Title: "Lord of the Rings: Return of the King", Quantity: 10},
		}

		existingUsers := []user.User{
			{FirstName: "Tine", LastName: "Kokalj"},
			{FirstName: "Žan", LastName: "Horvat"},
			{FirstName: "Luka", LastName: "Potočnik"},
		}

		test := struct {
			name     string
			input    book_borrow.BookBorrow
			expected int
		}{
			name:     "Return a book",
			input:    book_borrow.BookBorrow{BookID: 1, UserID: 1},
			expected: http.StatusOK,
		}

		for _, b := range existingBooks {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
			assert.NoError(t, err)
		}

		for _, u := range existingUsers {
			_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2)", u.FirstName, u.LastName)
			assert.NoError(t, err)
		}

		_, err := dbService.GetPool().Exec(context.Background(), "INSERT INTO book_borrows (book_id, user_id) VALUES (1, 1)")
		assert.NoError(t, err)
		_, err = dbService.GetPool().Exec(context.Background(), "INSERT INTO book_borrows (book_id, user_id) VALUES (2, 1)")
		assert.NoError(t, err)

		reqBody, err := json.Marshal(test.input)
		req := httptest.NewRequest("PUT", "/book_borrow", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, resp.StatusCode)

		req, _ = http.NewRequest("GET", "/book_borrowed", nil)
		resp, _ = app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var borrowedBooks []book_borrow.BookBorrow
		err = json.NewDecoder(resp.Body).Decode(&borrowedBooks)
		assert.NoError(t, err)
		assert.Len(t, borrowedBooks, 1)
	})
}
