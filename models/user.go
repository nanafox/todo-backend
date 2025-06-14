package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nanafox/todo-backend/config"
	"github.com/nanafox/todo-backend/utils"
)

var validate = validator.New()

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" validate:"required,email" gorm:"uniqueIndex;not null"`
	FirstName    string    `json:"first_name" validate:"required" gorm:"not null"`
	LastName     string    `json:"last_name" validate:"required" gorm:"not null"`
	PictureUrl   string    `json:"picture_url"`
	Password     string    `gorm:"-"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Tasks        []Task    `json:"tasks"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (user *User) Save() (err error) {
	if err = validate.Struct(user); err != nil {
		return err
	}

	user.PasswordHash, err = utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	result := config.DB.Create(&user)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (user *User) Name() string {
	return user.FirstName + " " + user.LastName
}
