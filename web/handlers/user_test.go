package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"kokal5296/database"
	"kokal5296/models/user"
	"kokal5296/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	connStr    = "postgres://postgres:postgres@localhost:5433/"
	testDbName = "test_db"
)

// SetupTestDB creates a new test database and returns a database service for it.
func SetupTestDB() (database.DatabaseService, func(), error) {
	dbService := database.NewDatabaseService()

	// Connect to the main "postgres" database for admin tasks
	adminConn, err := dbService.NewDatabase(connStr, "postgres")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to postgres database: %v", err)
	}
	defer adminConn.Close()

	// Drop existing test database if it exists
	_, err = adminConn.GetPool().Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s;", testDbName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to drop existing test database: %v", err)
	}

	conn, err := dbService.NewDatabase(connStr, testDbName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test database: %v", err)
	}

	teardown := func() {
		conn.Close()

		time.Sleep(100 * time.Millisecond)

		// Reconnect to the "postgres" database to terminate active connections and drop the test database
		dropConn, err := database.NewDatabaseService().NewDatabase(connStr, "postgres")
		if err != nil {
			fmt.Printf("Failed to connect to drop test database: %v\n", err)
			return
		}
		defer dropConn.Close()

		// Terminate active connections to the test database
		_, err = dropConn.GetPool().Exec(context.Background(), fmt.Sprintf(`
			SELECT pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = '%s' AND pid <> pg_backend_pid();`, testDbName))
		if err != nil {
			fmt.Printf("Failed to terminate connections to test database: %v\n", err)
			return
		}

		_, err = dropConn.GetPool().Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s;", testDbName))
		if err != nil {
			fmt.Printf("Failed to drop test database: %v\n", err)
		}
	}

	return dbService, teardown, nil
}

// TestCreateUser tests the scenarios for creating a new user.
func TestCreateUser(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	app := fiber.New()
	app.Post("/users", userApi.CreateUser)

	tests := []struct {
		name           string
		input          user.User
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Successful User Creation",
			input:          user.User{FirstName: "Tine", LastName: "Kokalj"},
			expectedStatus: http.StatusCreated,
			expectedCount:  1,
		},
		{
			name:           "Missing First Name",
			input:          user.User{LastName: "Kokalj"},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "Missing Last Name",
			input:          user.User{FirstName: "Tine"},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "Duplicate User",
			input:          user.User{FirstName: "Tine", LastName: "Kokalj"},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  1,
		},
		{
			name:           "Invalid Field Data",
			input:          user.User{FirstName: "", LastName: "Kokalj"},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/users", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var userCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE first_name = $1 AND last_name = $2", tt.input.FirstName, tt.input.LastName).Scan(&userCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, userCount)
		})
	}
}

// TestGetUser tests the scenarios for retrieving a user by ID.
func TestGetUser(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	app := fiber.New()
	app.Get("/users/:id", userApi.GetUser)

	existingUser := user.User{FirstName: "Tine", LastName: "Kokalj"}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2) RETURNING id", existingUser.FirstName, existingUser.LastName).Scan(&existingUser.ID)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		input          string
		expectedStatus int
		expectedUser   *user.User
	}{
		{
			name:           "Successful User Retrieval",
			input:          fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusOK,
			expectedUser:   &existingUser,
		},
		{
			name:           "User Not Found",
			input:          "9999",
			expectedStatus: http.StatusBadRequest,
			expectedUser:   nil,
		},
		{
			name:           "Invalid ID format",
			input:          "abc",
			expectedStatus: http.StatusBadRequest,
			expectedUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/users/"+tt.input, nil)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedUser != nil {
				var user user.User
				err = json.NewDecoder(resp.Body).Decode(&user)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, &user)
			}
		})
	}
}

// TestGetAllUsers tests the scenarios for retrieving all users.
func TestGetAllUsers(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	app := fiber.New()
	app.Get("/users", userApi.GetAllUsers)

	t.Run("Retrieve all users when users exist", func(t *testing.T) {
		existingUsers := []user.User{
			{FirstName: "Tine", LastName: "Kokalj"},
			{FirstName: "Gašper", LastName: "Zajc"},
			{FirstName: "Luka", LastName: "Potočnik"},
		}

		for _, u := range existingUsers {
			_, err = dbService.GetPool().Exec(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2)", u.FirstName, u.LastName)
			assert.NoError(t, err)
		}

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedUsers []user.User
		err = json.NewDecoder(resp.Body).Decode(&retrievedUsers)
		assert.NoError(t, err)
		assert.Len(t, retrievedUsers, len(existingUsers))

		for i, u := range existingUsers {
			assert.Equal(t, u.FirstName, retrievedUsers[i].FirstName)
			assert.Equal(t, u.LastName, retrievedUsers[i].LastName)
		}
	})

	t.Run("Retrieve all users when no users exist", func(t *testing.T) {

		_, err := dbService.GetPool().Exec(context.Background(), "DELETE FROM users")
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedUsers []user.User
		err = json.NewDecoder(resp.Body).Decode(&retrievedUsers)
		assert.NoError(t, err)
		assert.Empty(t, retrievedUsers)
	})
}

// TestUpdateUser tests the scenarios for updating a user.
func TestUpdateUser(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	app := fiber.New()
	app.Put("/users/:id", userApi.UpdateUser)

	existingUser := user.User{FirstName: "Tine", LastName: "Kokalj"}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2) RETURNING id", existingUser.FirstName, existingUser.LastName).Scan(&existingUser.ID)

	tests := []struct {
		name           string
		input          user.User
		id             string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Successful User Update",
			input:          user.User{FirstName: "Gašper", LastName: "Zajc"},
			id:             fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Duplicate User",
			input:          user.User{FirstName: "Gašper", LastName: "Zajc"},
			id:             fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  1,
		},
		{
			name:           "Missing First Name",
			input:          user.User{LastName: "Kokalj"},
			id:             fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusBadRequest,
			expectedCount:  1,
		},
		{
			name:           "Missing Last Name",
			input:          user.User{FirstName: "Tine"},
			id:             fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusBadRequest,
			expectedCount:  1,
		},
		{
			name:           "User Not Found",
			input:          user.User{FirstName: "Tine", LastName: "Kokalj"},
			id:             "9999",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  1,
		},
		{
			name:           "Invalid ID format",
			input:          user.User{FirstName: "Tine", LastName: "Kokalj"},
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("PUT", "/users/"+tt.id, bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var userCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM users ").Scan(&userCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, userCount)

		})
	}
}

// TestDeleteUser tests the scenarios for deleting a user.
func TestDeleteUser(t *testing.T) {

	dbService, teardown, err := SetupTestDB()
	assert.NoError(t, err)
	defer teardown()

	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	app := fiber.New()
	app.Delete("/users/:id", userApi.DeleteUser)

	existingUser := user.User{FirstName: "Tine", LastName: "Kokalj"}
	err = dbService.GetPool().QueryRow(context.Background(), "INSERT INTO users (first_name, last_name) VALUES ($1, $2) RETURNING id", existingUser.FirstName, existingUser.LastName).Scan(&existingUser.ID)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Invalid ID format",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  1,
		},
		{
			name:           "User Not Found",
			id:             "9999",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  1,
		},
		{
			name:           "Successful User Deletion",
			id:             fmt.Sprint(existingUser.ID),
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/users/"+tt.id, nil)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var userCount int
			err = dbService.GetPool().QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE id = $1", existingUser.ID).Scan(&userCount)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, userCount)
		})
	}
}
