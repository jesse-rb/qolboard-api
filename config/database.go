package config

import (
	"fmt"
	"os"
	"qolboard-api/api"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection *gorm.DB
}

func ConnectToDatabase() *Database {
	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logError.Panic(err)
	}

	db.AutoMigrate(&api.User{})

	return &Database{Connection: db}
}

func (db *Database) AutoMigrate(m interface{}) {
	err := db.Connection.AutoMigrate(&m)
	if err != nil {
		logError.Panic(err)
	}
}