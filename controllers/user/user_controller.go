package user_controller

import (
	"log"
	"net/http"
	"os"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger slogger.Logger = *slogger.New(os.Stdout, slogger.ANSIGreen, "user_controller", log.Lshortfile+log.Ldate)
var errorLogger slogger.Logger = *slogger.New(os.Stderr, slogger.ANSIRed, "user_controller", log.Lshortfile+log.Ldate)

func Get(c *gin.Context) {
	email, exists := c.Get("email")
	if (!exists) {
		response_service.SetCode(c, http.StatusUnauthorized)
		error_service.PublicError(c, "Could not find user", http.StatusUnauthorized, "auth", "", "user")
		return
	}

	response_service.SetJSON(c, gin.H{"email": email})
}
