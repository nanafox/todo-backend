package controllers

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nanafox/gofetch"
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

func GoogleOAuth(c *fiber.Ctx) error {
	var body struct {
		AccessToken string `json:"access_token"`
	}

	if err := c.BodyParser(&body); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// ✅ Fetch user info from Google
	client := gofetch.New(gofetch.Config{Timeout: 2 * time.Minute})
	header := gofetch.Header{
		Key:   "Authorization",
		Value: "Bearer " + body.AccessToken,
	}

	client.Get("https://www.googleapis.com/oauth2/v3/userinfo", nil, header)

	if client.Error != nil {
		log.Println("Error fetching user info from Google:", client.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":     false,
			"status_code": fiber.StatusInternalServerError,
			"message":     "Failed to fetch user info from Google",
		})
	}

	type UserInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"` // Google user ID
	}

	var userInfo UserInfo

	if err := client.ResponseToStruct(&userInfo); err != nil {
		log.Println("Error decoding user info:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":     false,
			"status_code": fiber.StatusInternalServerError,
			"message":     "Failed to decode user info",
		})
	}

	// ✅ Extract user info
	email := userInfo.Email
	name := userInfo.Name
	sub := userInfo.Sub

	log.Printf("Google user info: Email: %s, Name: %s, Sub: %s\n", email, name, sub)

	user := &models.User{}
	if err := user.FindByEmail(email); err != nil {
		nameSplit := strings.Split(name, " ")

		firstName := strings.Join(nameSplit[0:len(nameSplit)-2], " ")
		lastName := nameSplit[len(nameSplit)-1]

		user.Email = email
		user.FirstName = firstName
		user.LastName = lastName
		user.OAuthUser = true
		user.Password = "Password1234"

		log.Printf("User info: %+v\n", user)
		if err = user.Save(); err != nil {
			log.Println("Error saving user:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":     false,
				"status_code": fiber.StatusInternalServerError,
				"message":     "Failed to save user",
			})
		}
	}

	// ✅ Generate your own app's tokens
	tokens, err := utils.GenerateJWT(&utils.JWTClaims{
		Email: user.Email,
		Name:  user.Name(),
		Sub:   user.ID,
	}, utils.GenerateJWTOptions{AddNewRefreshToken: true})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":     false,
			"status_code": 500,
			"message":     "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  fiber.StatusOK,
		"message": "Google account verified",
		"data": fiber.Map{
			"access_token":  tokens["access_token"],
			"refresh_token": tokens["refresh_token"],
		},
	})
}

// Logout handles user logout by invalidating the session or token.
func Logout(c *fiber.Ctx) error {
	// This function will handle user logout
	// In a stateless application, logout is typically handled by deleting the token on the client side.
	// You can also implement token invalidation logic if needed.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
		"status":  fiber.StatusOK,
	})
}
