package response_middleware

import (
	"log"
	"os"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "response_middleware", log.Lshortfile+log.Ldate)
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "response_middleware", log.Lshortfile+log.Ldate)

func Run(c *gin.Context) {
	c.Next()

	var code int = response_service.GetCode(c)
	var response gin.H = response_service.GetJSON(c)

	c.JSON(code, response)
}