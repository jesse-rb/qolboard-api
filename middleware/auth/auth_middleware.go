package auth_middleware

import (
	"net/http"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	"qolboard-api/services/logging"

	"github.com/gin-gonic/gin"
)

// Authenticate middleware
func Run(c *gin.Context) {
	token, err := c.Cookie("qolboard_jwt")
	if err != nil {
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	if token == "" {
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	claims, err := auth_service.ParseJWT(token)
	if err != nil {
		logging.LogInfo("AuthMiddleware", "Error parsing token", err)
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	logging.LogInfo("AuthMiddleware", "Received request from", claims.Email)

	c.Set("claims", claims)
	c.Set("token", token)
	c.Next()
}
