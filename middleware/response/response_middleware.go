package response_middleware

import (
	"qolboard-api/services/logging"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

func Run(c *gin.Context) {
	c.Next()

	logging.LogDebug("[middleware]", "[response]", nil)
	if !c.Writer.Written() {
		response_service.Response(c)
	}
}
