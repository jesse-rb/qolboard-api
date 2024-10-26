package user_controller

import (
	"log"
	"os"
	auth_service "qolboard-api/services/auth"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger slogger.Logger = *slogger.New(os.Stdout, slogger.ANSIGreen, "user_controller", log.Lshortfile+log.Ldate)
var errorLogger slogger.Logger = *slogger.New(os.Stderr, slogger.ANSIRed, "user_controller", log.Lshortfile+log.Ldate)

func Get(c *gin.Context) {
	claims := auth_service.GetClaims(c)
	email := claims.Email
	
	response_service.SetJSON(c, gin.H{"email": email})
}
