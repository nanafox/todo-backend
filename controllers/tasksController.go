package controllers

import (
	"github.com/gofiber/fiber/v2"
)

func ListAllTasks(c *fiber.Ctx) error {
	tasks := make([]any, 0)

	tasks = append(tasks, fiber.Map{"id": 1, "name": "Get something done"})
	return c.Status(fiber.StatusOK).JSON(tasks)
}
