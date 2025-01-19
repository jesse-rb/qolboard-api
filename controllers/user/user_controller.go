package user_controller

import (
	auth_service "qolboard-api/services/auth"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context) {
	claims := auth_service.GetClaims(c)
	email := claims.Email

	response_service.SetJSON(c, gin.H{"email": email})
}
