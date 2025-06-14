package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB = DBConfig()

func DBConfig() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("big-guys-todo-app.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}

	return db
}
