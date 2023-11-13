package database

import (
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

var db *gorm.DB

func ConnectDB(filename string) {
	database, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = database
}

func GetDB() *gorm.DB {
	return db
}
