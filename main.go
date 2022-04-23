package main

import (
	"log"
	"os"
	"qolboard-api/api"
	"qolboard-api/config"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	
}

func main() {
	logError := log.New(os.Stdout, "main\t\t=> error\t\t=> ", log.LstdFlags)
	logInfo := log.New(os.Stdout, "main\t\t=> info\t\t=> ", log.LstdFlags)
	var err error;

	err = godotenv.Load()
	if err != nil {
		logError.Fatal("Error loading .env file")
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
	logInfo.Println("Running server")
	if err != nil {
		logError.Fatal("Error running server")
	}
}