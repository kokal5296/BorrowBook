package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	er "kokal5296/errors"
	"kokal5296/models/user"
	"kokal5296/service"
	validate "kokal5296/web/validation"
	"log"
	"net/http"
	"strconv"
)

type UserApiStruct struct {
	userService service.UserService
}

// NewUserApiService creates a new instance of UserApiStruct, which implements the UserApi interface
func NewUserApiService(userService service.UserService) UserApi {
	return &UserApiStruct{
		userService: userService,
	}
}

// CreateUser handles the request to create a new user
func (s *UserApiStruct) CreateUser(c *fiber.Ctx) error {

	log.Println("Requesting to create new user")
	var newUser user.User

	funcName := handler + "CreateUser"

	err := json.Unmarshal(c.Body(), &newUser)
	if err != nil {
		log.Printf("Error while unmarshalling user: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateUser(newUser)
	if validateErr != nil {
		log.Printf("Error while validating user: %v", validateErr)
		return c.Status(fiber.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.userService.CreateUser(c.Context(), newUser)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusCreated).SendString("User was successfully created")
}

// GetUser handles the request to get a user by id
func (s *UserApiStruct) GetUser(c *fiber.Ctx) error {

	log.Println("Requesting to get user by id")
	funcName := handler + "GetUser"
	id := c.Params("id")

	userId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Error while converting id to int: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	user, err := s.userService.GetUser(c.Context(), userId)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusBadRequest).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).JSON(user)
}

// GetAllUsers handles the request to get all users
func (s *UserApiStruct) GetAllUsers(c *fiber.Ctx) error {

	log.Println("Requesting to get all users")
	funcName := handler + "GetAllUsers"

	users, err := s.userService.GetAllUsers(c.Context())
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusBadRequest).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).JSON(users)
}

// UpdateUser handles the request to update a user
func (s *UserApiStruct) UpdateUser(c *fiber.Ctx) error {

	log.Println("Requesting to update user")
	var updateUser user.User
	funcName := handler + "UpdateUser"

	id := c.Params("id")

	userId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	err = json.Unmarshal(c.Body(), &updateUser)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	validateErr := validate.ValidateUser(updateUser)
	if validateErr != nil {
		return c.Status(http.StatusBadRequest).SendString(validateErr.Error())
	}

	err = s.userService.UpdateUser(c.Context(), updateUser, userId)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).SendString("User was updated successfully")
}

// DeleteUser handles the request to delete a user
func (s *UserApiStruct) DeleteUser(c *fiber.Ctx) error {

	log.Println("Requesting to delete user")
	id := c.Params("id")
	funcName := handler + "DeleteUser"

	userId, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	err = s.userService.DeleteUser(c.Context(), userId)
	if err != nil {
		er.Wrap(funcName, err)
		return c.Status(fiber.StatusInternalServerError).SendString(er.UnwrapError(err).Error())
	}

	return c.Status(http.StatusOK).SendString("User was successfully deleted")
}
