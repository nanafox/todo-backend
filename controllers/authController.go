package controllers

import (
	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "token": "abc123"})
}
