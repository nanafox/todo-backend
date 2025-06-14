package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/nanafox/todo-backend/config"
	"github.com/nanafox/todo-backend/controllers"
	"github.com/nanafox/todo-backend/models"
)

func main() {
	app := fiber.New(fiber.Config{AppName: "Big Guys Todo Backend"})
	app.Use(logger.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")
	auth := v1.Group("/auth")
	tasks := v1.Group("/tasks")

	config.DB.AutoMigrate(
		&models.User{}, &models.Task{}, &models.AccountIdentity{},
	)

	tasks.Get("/", controllers.ListAllTasks)
	auth.Post("/login", controllers.Login)
	auth.Post("/register", controllers.Register)

	log.Fatal(app.Listen(":3000"))
}
