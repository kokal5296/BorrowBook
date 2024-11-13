package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"kokal5296/models/book"
	"kokal5296/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCreateBook tests the scenarios for creating a new book
func TestCreateBook(t *testing.T) {
	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	bookService := service.NewBookService(dbService)
	bookApi := NewBookApiService(bookService)

	app := fiber.New()
	app.Post("/book", bookApi.CreateBook)

	tests := []struct {
		name               string
		input              book.Book
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "Create a new book",
			input:              book.Book{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 2},
			expectedStatusCode: fiber.StatusCreated,
			expectedCount:      1,
		},
		{
			name:               "Create a new book with empty title",
			input:              book.Book{Title: "", Quantity: 2},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedCount:      1,
		},
		{
			name:               "Create a new book with empty quantity",
			input:              book.Book{Title: "The Alchemist", Quantity: 0},
			expectedStatusCode: fiber.StatusBadRequest,
			expectedCount:      1,
		},
		{
			name:               "Create a new book with duplicate title",
			input:              book.Book{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 1},
			expectedStatusCode: fiber.StatusInternalServerError,
			expectedCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/book", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			var bookCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM books").Scan(&bookCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, bookCount)
		})
	}
}

// TestGetBook tests the scenarios for retrieving a book by ID
func TestGetBook(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	bookService := service.NewBookService(dbService)
	bookApi := NewBookApiService(bookService)

	app := fiber.New()
	app.Get("/book/:id", bookApi.GetBook)

	existingBook := book.Book{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 5}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2) RETURNING id", existingBook.Title, existingBook.Quantity).Scan(&existingBook.ID)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		input              string
		expectedStatusCode int
		expectedBook       *book.Book
	}{
		{
			name:               "Successful Book Retrieval",
			input:              fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusOK,
			expectedBook:       &existingBook,
		},
		{
			name:               "Book Not Found",
			input:              "100",
			expectedStatusCode: http.StatusInternalServerError,
			expectedBook:       nil,
		},
		{
			name:               "Invalid Book ID",
			input:              "invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedBook:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/book/"+tt.input, nil)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			if tt.expectedBook != nil {
				var book book.Book
				err = json.NewDecoder(resp.Body).Decode(&book)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBook, &book)
			}
		})
	}
}

// TestGetAllBooks tests the scenarios for retrieving all books
func TestGetAllBooks(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	bookService := service.NewBookService(dbService)
	bookApi := NewBookApiService(bookService)

	app := fiber.New()
	app.Get("/books", bookApi.GetAllBooks)

	t.Run("Retrieve all books when books exist", func(t *testing.T) {
		existingBooks := []book.Book{
			{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 5},
			{Title: "The Lord Of The Rings: The two Towers", Quantity: 3},
			{Title: "The Lord Of The Rings: The Return of the King", Quantity: 2},
		}

		for _, b := range existingBooks {
			_, err = dbService.GetPool().Exec(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2)", b.Title, b.Quantity)
			assert.NoError(t, err)
		}

		req := httptest.NewRequest("GET", "/books", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var books []book.Book
		err = json.NewDecoder(resp.Body).Decode(&books)
		assert.NoError(t, err)
		assert.Len(t, existingBooks, len(books))

		for i, u := range existingBooks {
			assert.Equal(t, u.Title, books[i].Title)
			assert.Equal(t, u.Quantity, books[i].Quantity)
		}
	})

	t.Run("Retrieve all books when no book exist", func(t *testing.T) {
		_, err = dbService.GetPool().Exec(context.Background(), "DELETE FROM books")
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/books", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var books []book.Book
		err = json.NewDecoder(resp.Body).Decode(&books)
		assert.NoError(t, err)
		assert.Empty(t, books)
	})
}

// TestUpdateBook tests the scenarios for updating a book by ID
func TestUpdateBook(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	bookService := service.NewBookService(dbService)
	bookApi := NewBookApiService(bookService)

	app := fiber.New()
	app.Put("/book/:id", bookApi.UpdateBook)

	existingBook := book.Book{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 5}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2) RETURNING id", existingBook.Title, existingBook.Quantity).Scan(&existingBook.ID)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		input              book.Book
		id                 string
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "Update book title and quantity",
			input:              book.Book{ID: existingBook.ID, Title: "The Lord Of The Rings: Return of the King", Quantity: 10},
			id:                 fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "Update book quantity",
			input:              book.Book{ID: existingBook.ID, Title: "The Lord Of The Rings: Return of the King", Quantity: 5},
			id:                 fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusOK,
			expectedCount:      1,
		},
		{
			name:               "Update book with empty title",
			input:              book.Book{ID: existingBook.ID, Title: "", Quantity: 5},
			id:                 fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      1,
		},
		{
			name:               "Update book with empty quantity",
			input:              book.Book{ID: existingBook.ID, Title: "The Lord Of The Rings: Return of the King"},
			id:                 fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      1,
		},
		{
			name:               "Update book with invalid id",
			input:              book.Book{Title: "The Lord Of The Rings: Return of the King", Quantity: 5},
			id:                 "100",
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      1,
		},
		{
			name:               "Update book with invalid id format",
			input:              book.Book{Title: "The Lord Of The Rings: Return of the King", Quantity: 5},
			id:                 "invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("PUT", "/book/"+tt.id, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			var bookCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM books").Scan(&bookCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, bookCount)
		})
	}
}

// TestDeleteBook tests the scenarios for deleting a book by ID
func TestDeleteBook(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	bookService := service.NewBookService(dbService)
	bookApi := NewBookApiService(bookService)

	app := fiber.New()
	app.Delete("/book/:id", bookApi.DeleteBook)

	existingBook := book.Book{Title: "The Lord Of The Rings: Fellowship of the Ring", Quantity: 5}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO books (title, quantity) VALUES ($1, $2) RETURNING id", existingBook.Title, existingBook.Quantity).Scan(&existingBook.ID)
	assert.NoError(t, err)

	tests := []struct {
		name               string
		id                 string
		expectedStatusCode int
		expectedCount      int
	}{
		{
			name:               "Delete book with invalid id",
			id:                 "100",
			expectedStatusCode: http.StatusInternalServerError,
			expectedCount:      1,
		},
		{
			name:               "Delete book with invalid id format",
			id:                 "invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedCount:      1,
		},
		{
			name:               "Delete book",
			id:                 fmt.Sprint(existingBook.ID),
			expectedStatusCode: http.StatusOK,
			expectedCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/book/"+tt.id, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			var bookCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM books").Scan(&bookCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, bookCount)
		})
	}
}
