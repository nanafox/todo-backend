package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/nanafox/todo-backend/errors"
	"github.com/nanafox/todo-backend/models"
)

func Register(c *fiber.Ctx) error {
	user := &models.User{}

	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"status":  fiber.StatusBadRequest,
		})
	}

	if err := user.Save(); err != nil {
		status := fiber.StatusBadRequest
		message := "Invalid request"

		switch {
		case errors.Is(err, customErrors.ErrDuplicateEmail):
			status = fiber.StatusConflict
			message = err.Error()
		case errors.Is(err, customErrors.ErrPasswordTooShort),
			errors.Is(err, customErrors.ErrPasswordTooWeak),
			errors.Is(err, customErrors.ErrInvalidEmail):
			message = err.Error()
		case errors.Is(err, customErrors.ErrInternal):
			status = fiber.StatusInternalServerError
			message = "Internal server error"
		default:
			var validationErr *customErrors.ValidationError
			if errors.As(err, &validationErr) {
				return c.Status(status).JSON(fiber.Map{
					"success": false,
					"field":   validationErr.Field,
					"message": validationErr.Message,
					"status":  status,
				})
			}
			// For unexpected errors
			message = err.Error()
		}

		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"message": message,
			"status":  status,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"status":  fiber.StatusCreated,
		"message": "User created successfully",
		"data": fiber.Map{
			"user_id":    user.ID,
			"name":       user.Name(),
			"created_at": user.CreatedAt,
		},
	})
}

func Login(c *fiber.Ctx) (err error) {
	credentials := &models.LoginUser{}

	if err = c.BodyParser(credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"status":  fiber.StatusBadRequest,
		})
	}

	user := &models.User{}
	if err = user.FindByEmail(credentials.Email); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid credentials",
			"status":  fiber.StatusUnauthorized,
		})
	}

	user.Password = credentials.Password
	authenticated, err := user.Authenticate()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
			"status":  fiber.StatusUnauthorized,
		})
	}

	if !authenticated {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid credentials",
			"status":  fiber.StatusUnauthorized,
		})
	}

	// TODO: Generate and return JWT token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"token":   "abc123",
		"status":  fiber.StatusOK,
	})
}
