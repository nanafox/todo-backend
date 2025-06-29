package errors

import "github.com/gofiber/fiber/v2"

func HandleUnauthorizedError(error error) fiber.Map {
	return fiber.Map{
		"success": false,
		"error":   error.Error(),
		"status":  fiber.StatusUnauthorized,
	}
}
