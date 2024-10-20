package database_config

import (
	"fmt"
	"log"
	"os"
	canvas_model "qolboard-api/models/canvas"

	slogger "github.com/jesse-rb/slogger-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declare some loggers
var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "database_config", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "database_config", log.Lshortfile+log.Ldate);

var database *Database

func GetDatabase() *Database {
	return database;
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
		errorLogger.Log("ConnectToDatabase", "Error connecting to database", err.Error())
		panic(1)
	}

	// Register auto migrations here:
	// e.g. db.AutoMigrate(&canvas_model.Canvas{})
	db.AutoMigrate(&canvas_model.Canvas{})

	database = &Database{Connection: db}
}

func (db *Database) AutoMigrate(m interface{}) {
	err := db.Connection.AutoMigrate(&m)
	if err != nil {
		errorLogger.Log("AutoMigrate", "Error auto migrating Gorm model", err)
	}
}