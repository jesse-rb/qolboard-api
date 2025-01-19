package main

import (
	"os"
	"qolboard-api/services/logging"

	database_config "qolboard-api/config/database"
	auth_controller "qolboard-api/controllers/auth"
	canvas_controller "qolboard-api/controllers/canvas"
	user_controller "qolboard-api/controllers/user"
	auth_middleware "qolboard-api/middleware/auth"
	cors_middleware "qolboard-api/middleware/cors"
	error_middleware "qolboard-api/middleware/error"
	response_middleware "qolboard-api/middleware/response"
	error_service "qolboard-api/services/error"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		logging.LogError("main", "Error loading .env file", err)
	}

	database_config.ConnectToDatabase()
}

func main() {
	// Setup router
	r := gin.Default()

	error_service.SetUpValidator()

	// Global middleware

	// Runs before
	r.Use(cors_middleware.Run)

	// Runs after (define in reverse)
	r.Use(response_middleware.Run)
	r.Use(error_middleware.Run)

	// Define unauthenticated routes routes
	// Auth routes
	rAuth := r.Group("/auth")
	{
		rAuth.POST("/register", auth_controller.Register)
		rAuth.POST("/login", auth_controller.Login)
		rAuth.POST("/set_token", auth_controller.SetToken)
		rAuth.POST("/resend_verification_email", auth_controller.ResendVerificationEmail)
	}

	// Define authenticated routes
	// User routes
	rUser := r.Group("/user")
	{
		// User middleware
		rUser.Use(auth_middleware.Run)

		rUser.GET("", user_controller.Get)
		rUser.POST("logout", auth_controller.Logout)

		// User Canvas routes
		rUser.POST("/canvas", canvas_controller.Save)
		rUser.GET("/canvas", canvas_controller.Index)
		rUser.GET("/canvas/:id", canvas_controller.Get)
		rUser.POST("/canvas/:id", canvas_controller.Save)
		rUser.DELETE("/canvas/:id", canvas_controller.Delete)

		rUser.GET("/ws/canvas/:id", canvas_controller.Websocket)
	}

	// Listen and serve router
	err := r.Run()
	logging.LogInfo("main", "Running server", 0)
	if err != nil {
		logging.LogError("main", "Error running server", err)
		os.Exit(1)
	}
}
