package main

import (
	"os"
	"qolboard-api/api"
	"qolboard-api/config"
	"qolboard-api/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var prefix string = "main"

func init() {
	
}

func main() {
	var err error;

	err = godotenv.Load()
	if err != nil {
		logger.LogError(prefix, "Error loading .env file");
	}

	// Setup router
	r := gin.Default();
	r.SetTrustedProxies([]string{os.Getenv("SPA_DOMAIN")})

	// Create database connection
	db := config.ConnectToDatabase()

	// Firebase auth connection
	firebaseAuth := config.FirebaseAuth()
	
	// Gin middleware for all requests
	r.Use(func(c *gin.Context) {
		c.Set("db", db.Connection)
		c.Set("firebaseAuth", firebaseAuth)
	})

	// Define routes
	r.GET("/user", api.GetUsers)
	r.POST("/user", api.CreateUser)

	// Listen and serve router
	err = r.Run()
	logger.LogInfo(prefix, "Server running")
	if err != nil {
		logger.LogError(prefix, "Error running server")
	}
}