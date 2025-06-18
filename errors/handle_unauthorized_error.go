package errors

import "github.com/gofiber/fiber/v2"

func HandleUnauthorizedError(message string) fiber.Map {
	return fiber.Map{
		"success": false,
		"message": message,
		"status":  fiber.StatusUnauthorized,
	}
}
