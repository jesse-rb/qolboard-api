package auth_controller

import (
	model "qolboard-api/models"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	supabase_service "qolboard-api/services/supabase"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var data supabase_service.RegisterBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	if data.Password != data.PasswordConfirmation {
		error_service.PublicError(c, "password confirmation does not match", 422, "password_confirmation", "", "user")
		return
	}

	code, response, err := supabase_service.Signup(data)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	if code != 200 {
		error_service.PublicError(c, response.Msg, code, "email", response.ErrorCode, "user")
	}

	user := model.User{
		Email: response.Email,
		Uuid:  response.Uuid,
	}

	response_service.SetJSON(c, gin.H{
		"data": user,
		"code": response.ErrorCode,
	})
}

func SetToken(c *gin.Context) {
	var data supabase_service.SetTokenBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	claims, err := auth_service.ParseJWT(data.Token)
	if err != nil {
		error_service.PublicError(c, "Invalid token", 401, "token", "", "")
		return
	}

	auth_service.SetAuthCookie(c, data.Token, data.ExpiresIn)

	user := model.User{
		Uuid:  claims.Subject,
		Email: claims.Email,
	}
	response_service.SetJSON(c, gin.H{"data": user})
}

func ResendVerificationEmail(c *gin.Context) {
	var data supabase_service.ResendEmailVerificationBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	// TODO: verify we have an unverified user for this email

	code, err := supabase_service.ResendEmailVerification(data.Email)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	if code != 200 {
		error_service.PublicError(c, "", code, "", "", "credentials")
		return
	}

	response_service.SetJSON(c, gin.H{"email": data.Email})
}

func Login(c *gin.Context) {
	var data supabase_service.LoginBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	code, response, err := supabase_service.Login(data)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	if code != 200 {
		error_service.PublicError(c, response.Msg, 401, "", "", "credentials")
		return
	}

	var token string = response.AccessToken
	var expiresIn int = response.ExpiresIn

	auth_service.SetAuthCookie(c, token, expiresIn)

	response_service.SetJSON(c, gin.H{"data": response.User})
}

func Logout(c *gin.Context) {
	code, err := supabase_service.Logout(c.GetString("token"))
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	if code < 200 && code >= 300 {
		error_service.PublicError(c, "Could not logout", 401, "", "", "")
		return
	}

	auth_service.ExpireAuthCookie(c)
}
