package models

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nanafox/todo-backend/config"
	customErrors "github.com/nanafox/todo-backend/errors"
	"github.com/nanafox/todo-backend/utils"
)

var validate = validator.New()

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" validate:"required,email" gorm:"uniqueIndex;not null"`
	FirstName    string    `json:"first_name" validate:"required" gorm:"not null"`
	LastName     string    `json:"last_name" validate:"required" gorm:"not null"`
	PictureUrl   string    `json:"picture_url"`
	Password     string    `gorm:"-" json:"password" validate:"required,min=8,max=15"`
	PasswordHash string    `json:"-" gorm:"not null"`
	OAuthUser    bool      `gorm:"-" json:"oauth_user"`
	Tasks        []Task    `json:"tasks"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (user *User) FindById(id uint) (err error) {
	result := config.DB.Where("users.id = ?", id).First(user)

	if result.Error != nil {
		log.Printf("User with ID: %v not found\n", id)
		return result.Error
	}

	log.Printf("User with ID %v found\n", id)
	return nil
}

func (user *User) FindByEmail(email string) (err error) {
	result := config.DB.Where("users.email = ?", email).First(user)

	if result.Error != nil {
		log.Printf("User with email: %v not found\n", email)
		return result.Error
	}

	log.Println("User found")
	return nil
}

func (user *User) Authenticate() (bool, error) {
	if user.PasswordHash == "" || user.Password == "" {
		return false, customErrors.ErrInvalidCredentials
	}

	if utils.VerifyPasswordHash(user.Password, user.PasswordHash) {
		return true, nil
	}

	return false, customErrors.ErrInvalidCredentials
}

func (user *User) Save() (err error) {
	// Validate struct
	if err = validate.Struct(user); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				return handleValidationError(e)
			}
		}
		return err
	}

	// Validate password
	if !user.OAuthUser {
		if err := validatePassword(user.Password); err != nil {
			return err
		}

		// Hash password
		user.PasswordHash, err = utils.HashPassword(user.Password)
		if err != nil {
			return customErrors.ErrInternal
		}
	}

	// Try to save
	result := config.DB.Create(&user)
	if result.Error != nil {
		log.Println(result.Error)
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed: users.email") {
			return customErrors.ErrDuplicateEmail
		}
		return customErrors.ErrInternal
	}

	return nil
}

func handleValidationError(e validator.FieldError) error {
	field := strings.ToLower(e.Field())

	switch {
	case e.Tag() == "required":
		return customErrors.NewValidationError(field, "this field is required")
	case e.Tag() == "email":
		return customErrors.ErrInvalidEmail
	default:
		return customErrors.NewValidationError(field, "invalid value")
	}
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return customErrors.ErrPasswordTooShort
	}

	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return customErrors.ErrPasswordTooWeak
	}

	return nil
}

func (user *User) Name() string {
	return user.FirstName + " " + user.LastName
}
