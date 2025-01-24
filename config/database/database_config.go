package database_config

import (
	"fmt"
	"os"
	model "qolboard-api/models"
	"qolboard-api/services/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var database *Database

func GetDatabase() *Database {
	return database
}

type Database struct {
	Connection *gorm.DB
}

func ConnectToDatabase() {
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
		logging.LogError("ConnectToDatabase", "Error connecting to database", err.Error())
		panic(1)
	}

	// Register auto migrations here:
	// e.g. db.AutoMigrate(&model.Canvas{})
	db.AutoMigrate(&model.Canvas{})
	db.AutoMigrate(&model.CanvasSharedInvitation{})
	db.AutoMigrate(&model.CanvasSharedAccess{})

	database = &Database{Connection: db}
}

func (db *Database) AutoMigrate(m interface{}) {
	err := db.Connection.AutoMigrate(&m)
	if err != nil {
		logging.LogError("AutoMigrate", "Error auto migrating Gorm model", err)
	}
}
