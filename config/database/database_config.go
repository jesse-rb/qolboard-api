package database_config

import (
	"fmt"
	"os"
	auth_service "qolboard-api/services/auth"
	"qolboard-api/services/logging"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func DB(c *gin.Context) (*sqlx.Tx, error) {
	user_uuid := auth_service.GetClaims(c).Subject

	// Begin transaction
	tx, err := db.Beginx()
	if err != nil {
		logging.LogError("DB", "Failed to being database transaction", err.Error())
		return nil, err
	}

	// Set the required databse session variables for the transaction, for RLS purposes
	_, err = tx.Exec(fmt.Sprintf("SET myapp.user_uuid = '%s'", user_uuid))
	if err != nil {
		tx.Rollback()
		logging.LogError("DB", "Failed to SET databse session user_uuid REQUIRED for RLS", err.Error())
		return nil, err
	}

	_, err = tx.Exec("SET ROLE anon")
	if err != nil {
		tx.Rollback()
		logging.LogError("DB", "Failed to SET databse session role REQUIRED for RLS", err.Error())
		return nil, err
	}

	return tx, err
}

func ConnectToDatabase() {
	var err error = nil

	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
	)

	db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		logging.LogError("ConnectToDatabase", "Error connecting to database", err.Error())
		panic(1)
	}
}
