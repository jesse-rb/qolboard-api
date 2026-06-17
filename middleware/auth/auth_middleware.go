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
	unauthorized := false

	token, err := auth_service.GetJWTCookie(c)
	if err != nil {
		unauthorized = true
	}

	if token == "" {
		unauthorized = true
	}

	claims, err := auth_service.ParseJWT(token)
	if err != nil {
		unauthorized = true
		logging.LogDebug("AuthMiddleware", "Error parsing token", err)
	}

	if unauthorized {
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	logging.LogInfo("AuthMiddleware", "Received request from", claims.Email)

	c.Set("claims", claims)
	c.Set("token", token)
	c.Next()
}
