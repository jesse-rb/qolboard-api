package cors_middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Run(c *gin.Context) {
	var appHost string = os.Getenv("APP_HOST")

	c.Writer.Header().Set("Access-Control-Allow-Origin", appHost)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusContinue)
	}

	c.Next()
}
