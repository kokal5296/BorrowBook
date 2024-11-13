package service

import (
	"context"
	"fmt"
	"kokal5296/database"
	er "kokal5296/errors"
	"kokal5296/models/user"
	"log"
	"time"
)

type UserServiceStruct struct {
	dbService database.DatabaseService
}

const userService = "userService - "

// UserService interface defines methods for user-related operations
type UserService interface {
	CreateUser(ctx context.Context, newUser user.User) error
	GetUser(ctx context.Context, userId int) (*user.User, error)
	GetAllUsers(ctx context.Context) ([]user.User, error)
	UpdateUser(ctx context.Context, user user.User, userId int) error
	DeleteUser(ctx context.Context, userId int) error
	UserExist(ctx context.Context, userId int) error
}

// NewUserService creates a new instance of UserServiceStruct, implementing UserService
func NewUserService(dbService database.DatabaseService) UserService {
	return &UserServiceStruct{
		dbService: dbService,
	}
}

// CreateUser creates a new user in the database
func (s *UserServiceStruct) CreateUser(ctx context.Context, newUser user.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := userService + "CreateUser,"

	err := s.nameAndLastNameExist(ctx, newUser)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	query := `INSERT INTO users (first_name, last_name) VALUES ($1, $2)`
	_, err = s.dbService.GetPool().Exec(ctx, query, newUser.FirstName, newUser.LastName)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return er.Wrap(funcName, err)
		}
		log.Printf("Error creating user: %v", err)
		return er.Wrap(funcName, err)
	}

	log.Println("User created")
	return nil
}

// GetUser retrieves a user from the database by their ID
func (s *UserServiceStruct) GetUser(ctx context.Context, userId int) (*user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := userService + "GetUser,"

	var user user.User
	query := `SELECT id,  first_name, last_name FROM users WHERE id = $1`
	err := s.dbService.GetPool().QueryRow(ctx, query, userId).Scan(&user.ID, &user.FirstName, &user.LastName)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		log.Printf("Error getting user: %v", err)
		return nil, er.Wrap(funcName, err)
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database
func (s *UserServiceStruct) GetAllUsers(ctx context.Context) ([]user.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := userService + "GetAllUsers,"

	var users []user.User
	query := `SELECT id,  first_name, last_name FROM users`
	rows, err := s.dbService.GetPool().Query(ctx, query)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return nil, er.Wrap(funcName, err)
		}
		message := fmt.Sprintf("Error getting all users")
		return nil, er.New(funcName, message, err)
	}
	defer rows.Close()

	for rows.Next() {
		var user user.User
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName)
		if err != nil {
			message := fmt.Sprintf("Error scanning user: %v", err)
			return nil, er.New(funcName, message, err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates a user's information in the database
func (s *UserServiceStruct) UpdateUser(ctx context.Context, updateUser user.User, userId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := userService + "UpdateUser,"

	err := s.UserExist(ctx, userId)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	err = s.nameAndLastNameExist(ctx, updateUser)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	query := `UPDATE users SET first_name = $1, last_name = $2 WHERE id = $3`
	_, err = s.dbService.GetPool().Exec(ctx, query, updateUser.FirstName, updateUser.LastName, userId)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return er.Wrap(funcName, err)
		}
		message := fmt.Sprintf("Error updating user")
		return er.New(funcName, message, err)
	}

	return nil
}

// DeleteUser deletes a user from the database by their ID, if the user exists
func (s *UserServiceStruct) DeleteUser(ctx context.Context, userId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	funcName := userService + "DeleteUser,"

	err := s.UserExist(ctx, userId)
	if err != nil {
		return er.Wrap(funcName, err)
	}

	query := `DELETE FROM users WHERE id = $1`
	_, err = s.dbService.GetPool().Exec(ctx, query, userId)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return er.Wrap(funcName, err)
		}
		message := fmt.Sprintf("Error deleting user")
		return er.New(funcName, message, err)
	}

	return nil
}

// UserExist checks if a user with the given ID exists in the database
// This ensures that the user to be updated or deleted exists
func (s *UserServiceStruct) UserExist(ctx context.Context, userId int) error {

	funcName := userService + "userExist,"
	var userExists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`
	err := s.dbService.GetPool().QueryRow(ctx, query, userId).Scan(&userExists)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return er.Wrap(funcName, err)
		}
		message := fmt.Sprintf("Error checking if user exists")
		return er.New(funcName, message, err)
	}

	if !userExists {
		message := fmt.Sprintf("User with id %d does not exist", userId)
		return er.New(funcName, message, nil)
	}

	return nil
}

// nameAndLastNameExist checks if a user with the same first and last name already exists in the database
// This ensure that there are no duplicate users
func (s *UserServiceStruct) nameAndLastNameExist(ctx context.Context, user user.User) error {

	funcName := userService + "NameAndLastNameExist,"
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE first_name = $1 AND last_name = $2)`
	err := s.dbService.GetPool().QueryRow(ctx, query, user.FirstName, user.LastName).Scan(&exists)
	if err != nil {
		if er.HandleDeadlineExceededError(userService, err) != nil {
			return er.Wrap(funcName, err)
		}
		message := fmt.Sprintf("Error checking if user name and last name exists")
		return er.New(funcName, message, err)
	}

	if exists {
		message := fmt.Sprintf("User with this name: %s, and last name: %s, already exists", user.FirstName, user.LastName)
		return er.New(funcName, message, nil)
	}

	return nil
}
