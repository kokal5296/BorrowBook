package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	er "kokal5296/errors"
	"log"
)

type PostgreSQLConnection struct {
	Pool *pgxpool.Pool
}

const database = "database - "

// DatabaseService interface defines methods for database-related operations
type DatabaseService interface {
	NewDatabase(connStr string, dbName string) (*PostgreSQLConnection, error)
	Close()
	GetPool() *pgxpool.Pool
}

// NewDatabaseService creates a new instance of the PostgreSQLConnection struct, implementing the DatabaseService interface
func NewDatabaseService() DatabaseService {
	return &PostgreSQLConnection{}
}

// NewDatabase initializes a connection to PostgreSQL, checks if the target database exists,
// creates it if needed, and sets up the connection pool to the specific database.
func (db *PostgreSQLConnection) NewDatabase(connStr string, dbName string) (*PostgreSQLConnection, error) {

	funcName := database + "NewDatabase,"
	conn, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		message := fmt.Sprintf("Unable to connect to database")
		return nil, er.New(funcName, message, err)
	}
	defer conn.Close()

	var exists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM pg_database WHERE LOWER(datname) = LOWER($1))", dbName).Scan(&exists)
	if err != nil {
		message := fmt.Sprintf("Unable to check if database exists")
		return nil, er.New(funcName, message, err)
	}

	log.Printf("Database exists: %v\n", exists)

	if !exists {
		_, err = conn.Exec(context.Background(), "CREATE DATABASE "+dbName)
		if err != nil {
			message := fmt.Sprintf("Unable to create database")
			return nil, er.New(funcName, message, err)
		}
		log.Println("Database created")
	}

	finalConnStr := fmt.Sprintf("%s%s?sslmode=disable", connStr, dbName)
	log.Println("Connecting to database")

	pool, err := pgxpool.Connect(context.Background(), finalConnStr)
	if err != nil {
		message := fmt.Sprintf("Unable to connect to database")
		return nil, er.New(funcName, message, err)
	}
	db.Pool = pool

	log.Println("Database connection established")

	err = db.CreateTablesIfNotExist()
	if err != nil {
		message := fmt.Sprintf("Unable to create tables")
		return nil, er.New(funcName, message, err)
	}

	return &PostgreSQLConnection{Pool: pool}, nil
}

// CreateTablesIfNotExist creates tables if they do not exist
func (db *PostgreSQLConnection) CreateTablesIfNotExist() error {
	funcName := database + "CreateTablesIfNotExist,"

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            first_name VARCHAR(100) NOT NULL,
            last_name VARCHAR(100) NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS books (
            id SERIAL PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            quantity INT NOT NULL CHECK (quantity >= 0)
        );`,
		`CREATE TABLE IF NOT EXISTS book_borrows (
            id SERIAL PRIMARY KEY,
            user_id INT NOT NULL REFERENCES users(id),
            book_id INT NOT NULL REFERENCES books(id),
            borrow_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            return_date TIMESTAMP WITH TIME ZONE,
            CONSTRAINT unique_borrow UNIQUE(user_id, book_id, return_date)
        );`,
	}

	for _, query := range queries {
		_, err := db.Pool.Exec(context.Background(), query)
		if err != nil {
			message := fmt.Sprintf("Unable to create tables")
			return er.New(funcName, message, err)
		}
	}

	log.Println("Tables created or already exist")
	return nil
}

// Close closes the database connection
func (db *PostgreSQLConnection) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}

// GetPool returns the database connection pool
func (db *PostgreSQLConnection) GetPool() *pgxpool.Pool {
	return db.Pool
}
