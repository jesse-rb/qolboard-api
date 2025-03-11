package response_middleware

import (
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

func Run(c *gin.Context) {
	c.Next()

	response_service.Response(c)
}
