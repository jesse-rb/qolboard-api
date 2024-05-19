package database_middleware

import (
	database_config "qolboard-api/config/database"

	"github.com/gin-gonic/gin"
)

func Run(c *gin.Context) {
	// Create database connection and set in context
	db := database_config.ConnectToDatabase()
	c.Set("db", db.Connection)
}