package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nanafox/todo-backend/models"
)

func Register(c *fiber.Ctx) (err error) {
	user := &models.User{}

	if err = c.BodyParser(user); err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(
			fiber.Map{
				"success":     false,
				"message":     "Invalid request body",
				"status_code": fiber.ErrBadRequest.Code,
			},
		)
	}

	err = user.Save()

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"success":     false,
				"message":     "Invalid request body",
				"error":       err.Error(),
				"status_code": fiber.StatusBadRequest,
			},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		fiber.Map{
			"success":     true,
			"status_code": fiber.StatusCreated,
			"message":     "User created successfully",
			"data": map[string]any{
				"user_id":    user.ID,
				"created_at": user.CreatedAt,
			},
		},
	)
}

func Login(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "token": "abc123"})
}
