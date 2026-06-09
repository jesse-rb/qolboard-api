package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"qolboard-api/config"
	"qolboard-api/controllers"
	"qolboard-api/services/email"
	"qolboard-api/services/logging"
	"syscall"
	"time"

	database_config "qolboard-api/config/database"
	canvas_controller "qolboard-api/controllers/canvas"
	canvas_shared_access_controller "qolboard-api/controllers/canvas_shared_access"
	canvas_shared_invitation_controller "qolboard-api/controllers/canvas_shared_invitation"
	user_controller "qolboard-api/controllers/user"
	auth_middleware "qolboard-api/middleware/auth"
	cors_middleware "qolboard-api/middleware/cors"
	error_middleware "qolboard-api/middleware/error"
	rate_limiting_middleware "qolboard-api/middleware/rate_limiting"
	response_middleware "qolboard-api/middleware/response"
	error_service "qolboard-api/services/error"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Ensure our services are working in UTC
	os.Setenv("TZ", "UTC")

	err := godotenv.Load()
	if err != nil {
		logging.LogError("main", "Error loading .env file", err.Error())
	}

	database_config.ConnectToDatabase()
}

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Setup email client
	fromEmail := "info@qolboard.com"
	fromName := "qolboard"
	var emailCleint email.EmailClient = email.NewLogClient(fromEmail, fromName)
	if !config.IsDev() {
		var err error
		emailCleint, err = email.NewSESClient(ctx, fromEmail, fromName)
		if err != nil {
			logging.LogError("main", "error starting ses client", err)
			os.Exit(1)
		}
	}

	restHandler := controllers.NewRESTHAndler(emailCleint)

	// Setup router
	r := gin.Default()

	error_service.SetUpValidator()

	// Global middleware

	// Runs before
	r.Use(cors_middleware.Run)
	r.Use(rate_limiting_middleware.RunRateLimitIP(ctx))

	// Runs after (define in reverse)
	r.Use(response_middleware.Run)
	r.Use(error_middleware.Run)

	// Handle unregistered routes or methods
	r.NoRoute(func(c *gin.Context) {
		c.AbortWithError(404, fmt.Errorf("not found"))
	})
	// Define unauthenticated routes routes
	// Auth routes
	rAuth := r.Group("/auth")
	{
		rAuth.POST("/register", restHandler.Register)
		rAuth.GET("/verify", restHandler.VerifyEmail)
		rAuth.POST("/request_otp", restHandler.RequestOTP)
		rAuth.POST("/login", restHandler.Login)
	}

	// Define authenticated routes
	// User routes
	rUser := r.Group("/user")
	{
		// User middleware
		rUser.Use(auth_middleware.Run)
		rUser.Use(rate_limiting_middleware.RunRateLimitUser(ctx))

		rUser.GET("", user_controller.Get)
		rUser.POST("/logout", restHandler.Logout)

		// User Canvas routes
		rUser.POST("/canvas", canvas_controller.Save)
		rUser.GET("/canvas", canvas_controller.Index)
		rUser.GET("/canvas/:canvas_id", canvas_controller.Get)
		rUser.POST("/canvas/:canvas_id", canvas_controller.Save)
		rUser.DELETE("/canvas/:canvas_id", canvas_controller.Delete)

		rUser.GET("/canvas/:canvas_id/accept_invite/:code", canvas_shared_invitation_controller.AcceptInvite)

		rUser.POST("/canvas/:canvas_id/shared_invitation", canvas_shared_invitation_controller.Create)
		rUser.GET("/canvas/shared_invitation", canvas_shared_invitation_controller.Index)
		rUser.DELETE("/canvas/shared_invitation/:canvas_shared_invitation_id", canvas_shared_invitation_controller.Delete)
		//
		rUser.GET("/canvas/shared_access", canvas_shared_access_controller.Index)
		rUser.DELETE("/canvas/shared_access/:canvas_shared_access_id", canvas_shared_access_controller.Delete)
		//
		rUser.GET("/ws/canvas/:id", canvas_controller.Websocket)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: r,
	}

	// Listen and serve router
	logging.LogInfo("main", "Running server", 0)

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		var err error
		if config.IsDev() {
			// err = r.Run()
			err = srv.ListenAndServe()
		} else {
			// err = autotls.Run(r, os.Getenv("API_DOMAIN"))
			err = autotls.Run(srv.Handler, os.Getenv("API_DOMAIN"))
		}
		if err != nil && err != http.ErrServerClosed {
			logging.LogError("main", "Error running server", err)
			os.Exit(1)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logging.LogInfo("main", "shutting down gracefully, press Ctrl+C again to force", nil)

	// The context is used to inform the server it now has a timeout to finish any processing/handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logging.LogInfo("main", "server shutdown", err)
	}

	logging.LogInfo("main", "server finished, exiting", nil)
}
