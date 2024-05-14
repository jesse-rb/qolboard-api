package main

import (
	"log"
	"os"
	test_controller "qolboard-api/api/test"
	"qolboard-api/config"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
	"github.com/joho/godotenv"
)

func init() {
	
}

// Declare some loggers
var infoLogger = slogger.New(os.Stdout, slogger.ANSIBlue, "main", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "main", log.Lshortfile+log.Ldate);

func main() {
	var err error;

	err = godotenv.Load()
	if err != nil {
		errorLogger.Log("main", "Error loading .env file", err)
		os.Exit(1)
	}

	// Setup router
	r := gin.Default();
	r.SetTrustedProxies([]string{os.Getenv("SPA_DOMAIN")})

	// Create database connection
	db := config.ConnectToDatabase()
	
	// Gin middleware for all requests
	r.Use(func(c *gin.Context) {
		c.Set("db", db.Connection)
	})

	// Define routes
	r.GET("/test", test_controller.Index)

	// Listen and serve router
	err = r.Run()
	infoLogger.Log("main", "Running server", 0)
	if err != nil {
		errorLogger.Log("main", "Error running server", err)
		os.Exit(1)
	}
}