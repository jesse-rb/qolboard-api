package auth_middleware

import (
	"log"
	"net/http"
	"os"
	auth_service "qolboard-api/services/auth"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "auth_middleware", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "auth_middleware", log.Lshortfile+log.Ldate);

// Authenticate middleware
func Run(c *gin.Context) {
	token, err := c.Cookie("qolboard_jwt")
	
	if (err != nil) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	if (token == "") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	email, err := auth_service.ParseJWT(token)

	if (err != nil) {
		infoLogger.Log("AuthMiddleware", "Error parsing token", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	infoLogger.Log("AuthMiddleware", "Received request from", email)

	c.Set("email", email)
	c.Set("token", token)
	c.Next()
}
