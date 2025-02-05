package user_controller

import (
	model "qolboard-api/models"
	auth_service "qolboard-api/services/auth"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context) {
	claims := auth_service.GetClaims(c)

	var user model.User = model.User{
		Uuid:  claims.Subject,
		Email: claims.Email,
	}

	response_service.SetJSON(c, gin.H{
		"data": user,
	})
}
