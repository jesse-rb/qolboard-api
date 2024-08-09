package user_controller

import (
	"log"
	"net/http"
	"os"
	error_service "qolboard-api/services/error"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger slogger.Logger = *slogger.New(os.Stdout, slogger.ANSIGreen, "user_controller", log.Lshortfile+log.Ldate)
var errorLogger slogger.Logger = *slogger.New(os.Stderr, slogger.ANSIRed, "user_controller", log.Lshortfile+log.Ldate)

func Get(c *gin.Context) {
	email, exists := c.Get("email")
	if (!exists) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Could not find user"})
		error_service.PublicError(c, "Could not find user", http.StatusUnauthorized, "auth", "", "user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"email": email});
}
