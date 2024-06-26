package main

import (
	"log"
	"os"

	database_config "qolboard-api/config/database"
	auth_controller "qolboard-api/controllers/auth"
	user_controller "qolboard-api/controllers/user"
	auth_middleware "qolboard-api/middleware/auth"
	cors_middleware "qolboard-api/middleware/cors"
	database_middleware "qolboard-api/middleware/database"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		errorLogger.Log("main", "Error loading .env file", err)
		os.Exit(1)
	}

	database_config.ConnectToDatabase()
}

// Declare some loggers
var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "main", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "main", log.Lshortfile+log.Ldate);

func main() {
	// Setup router
	r := gin.Default();

	r.Use(cors_middleware.Run)

	// Global middleware
	r.Use(database_middleware.Run)

	// Define unauthenticated routes routes
	// Auth routes
	rAuth := r.Group("/auth")
	{
		rAuth.POST("/register", auth_controller.Register)
		rAuth.POST("/login", auth_controller.Login)
	}

	// Define authenticated routes
	// User routes
	rUser := r.Group("/user")
	{
		// User middleware
		rUser.Use(auth_middleware.Run)

		rUser.GET("", user_controller.Get)
		rUser.POST("logout", auth_controller.Logout)
	}


	// Listen and serve router
	err := r.Run()
	infoLogger.Log("main", "Running server", 0)
	if err != nil {
		errorLogger.Log("main", "Error running server", err)
		os.Exit(1)
	}
}