package response_middleware

import (
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

func Run(c *gin.Context) {
	c.Next()

	var code int = response_service.GetCode(c)
	var response gin.H = response_service.GetJSON(c)

	c.JSON(code, response)
}

