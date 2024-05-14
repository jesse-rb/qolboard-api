package config

import (
	"fmt"
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection *gorm.DB
}

// Declare some loggers
var infoLogger = slogger.New(os.Stdout, slogger.ANSIBlue, "config", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "config", log.Lshortfile+log.Ldate);

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
		errorLogger.Log("ConnectToDatabase", "Error connecting to database", err)
		panic(1)
	}

	// Register auto migrations here:
	// e.g. db.AutoMigrate(&api.User{})

	return &Database{Connection: db}
}

func (db *Database) AutoMigrate(m interface{}) {
	err := db.Connection.AutoMigrate(&m)
	if err != nil {
		errorLogger.Log("AutoMigrate", "Error auto migrating Gorm model", err)
	}
}