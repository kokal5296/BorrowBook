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

// Setup a new test database for each test
func setupTestDB() (database.DatabaseService, func(), error) {
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

	// Create a new test database
	conn, err := dbService.NewDatabase(connStr, testDbName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test database: %v", err)
	}

	// Define teardown function to drop the test database after the test
	teardown := func() {
		conn.Close()

		// Wait briefly to ensure all connections to the test DB are closed
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

		// Drop the test database
		_, err = dropConn.GetPool().Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s;", testDbName))
		if err != nil {
			fmt.Printf("Failed to drop test database: %v\n", err)
		}
	}

	return dbService, teardown, nil
}

func TestCreateUser(t *testing.T) {
	// Setup database and teardown after test
	dbService, teardown, err := setupTestDB()
	assert.NoError(t, err)
	defer teardown()

	// Initialize user service and API
	userService := service.NewUserService(dbService)
	userApi := NewUserApiService(userService)

	// Initialize Fiber app and add route
	app := fiber.New()
	app.Post("/users", userApi.CreateUser)

	// Prepare test data
	tests := []struct {
		name           string
		input          user.User
		expectedStatus int
		expectedCount  int // Expected user count in DB after test
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
