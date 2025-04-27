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

var (
	db     *sqlx.DB
	dbPriv *sqlx.DB
)

func DB(c *gin.Context) (*sqlx.Tx, error) {
	return beginDbTransaction(c)
}

func beginDbTransaction(c *gin.Context) (*sqlx.Tx, error) {
	user_uuid := auth_service.GetClaims(c).Subject

	// Begin transaction
	var tx *sqlx.Tx
	var err error

	tx, err = db.Beginx()
	if err != nil {
		logging.LogError("DB", "Failed to being database transaction", err.Error())
		return nil, err
	}

	// Set the required databse session variables for the transaction, for RLS purposes
	_, err = tx.Exec("SELECT set_user_uuid($1)", user_uuid)
	if err != nil {
		tx.Rollback()
		logging.LogError("DB", "Failed to SET databse session user_uuid REQUIRED for RLS", err.Error())
		return nil, err
	}

	return tx, err
}

func ConnectToDatabase() {
	var err error = nil

	host := os.Getenv("DB_HOST")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%s",
		host,
		username,
		name,
		password,
		port,
	)

	db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		logging.LogError("ConnectToDatabase", "Error connecting to database", err.Error())
		panic(1)
	}
}
