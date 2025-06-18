package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/nanafox/todo-backend/errors"
	"github.com/nanafox/todo-backend/models"
	"github.com/nanafox/todo-backend/utils"
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
		return c.Status(fiber.StatusUnauthorized).JSON(
			customErrors.HandleUnauthorizedError(customErrors.ErrInvalidCredentials.Error()),
		)
	}

	user.Password = credentials.Password
	authenticated, err := user.Authenticate()
	if err != nil || !authenticated {
		return c.Status(fiber.StatusUnauthorized).JSON(customErrors.HandleUnauthorizedError(err.Error()))
	}

	tokens, err := utils.GenerateJWT(&utils.JWTClaims{
		Sub:   user.ID,
		Email: user.Email,
		Name:  user.Name(),
	}, utils.GenerateJWTOptions{AddNewRefreshToken: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"success":     false,
				"message":     "Internal server error",
				"status_code": fiber.StatusInternalServerError,
			},
		)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  fiber.StatusOK,
		"message": "Log in successful",
		"data": fiber.Map{
			"access_token":  tokens["access_token"],
			"refresh_token": tokens["refresh_token"],
		},
	})
}

func RefreshToken(c *fiber.Ctx) error {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"status":  fiber.StatusBadRequest,
		})
	}

	refreshToken := req.RefreshToken

	userId, err := utils.VerifyJWT(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			customErrors.HandleUnauthorizedError("Invalid refresh token"),
		)
	}

	user := &models.User{}
	if err := user.FindById(userId); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			customErrors.HandleUnauthorizedError("Invalid refresh token or unauthenticated"),
		)
	}
	tokens, err := utils.GenerateJWT(&utils.JWTClaims{
		Sub:   user.ID,
		Email: user.Email,
		Name:  user.Name(),
	}, utils.GenerateJWTOptions{AddNewRefreshToken: false})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":     false,
			"message":     "Internal server error",
			"status_code": fiber.StatusInternalServerError,
		})
	}

	tokens["refresh_token"] = refreshToken // Keep the same refresh token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  fiber.StatusOK,
		"message": "Token refreshed successfully",
		"data": fiber.Map{
			"access_token":  tokens["access_token"],
			"refresh_token": tokens["refresh_token"],
		},
	})
}
