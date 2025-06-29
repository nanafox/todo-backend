package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/nanafox/todo-backend/config"
	"github.com/nanafox/todo-backend/controllers"
	"github.com/nanafox/todo-backend/models"
	"github.com/nanafox/todo-backend/utils"
)

func main() {
	app := fiber.New(fiber.Config{AppName: "Big Guys Todo Backend"})
	app.Use(logger.New())
	app.Use(cors.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	tasks := v1.Group("/tasks")

	err := config.DB.AutoMigrate(
		&models.User{}, &models.Task{}, &models.AccountIdentity{},
	)
	if err != nil {
		log.Fatal(err)
	}

	tasks.Use(utils.BearerTokenAuthenticationMiddleware())
	tasks.Get("/", controllers.ListAllTasks)

	// auth endpoints

	// api/v1/auth/login
	auth.Post("/login", controllers.Login)
	// api/v1/auth/register
	auth.Post("/register", controllers.Register)
	auth.Post("/refresh-token", controllers.RefreshToken)
	auth.Post("/google/callback", controllers.GoogleOAuth)

	auth.Use(utils.BearerTokenAuthenticationMiddleware())
	auth.Post("/logout", controllers.Logout)

	log.Fatal(app.Listen(":3000"))
}
