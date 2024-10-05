package auth_controller

import (
	"log"
	"net/http"
	"os"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	supabase_service "qolboard-api/services/supabase"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "auth_controller", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "auth_controller", log.Lshortfile+log.Ldate);

func Register(c *gin.Context) {
	var data supabase_service.RegisterBodyData

	err := c.ShouldBindJSON(&data)

	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	if data.Password != data.PasswordConfirmation {
		error_service.PublicError(c, "password confirmation does not match", 422, "password_confirmation", data.PasswordConfirmation, "user")
		return
	}

	code, response, err := supabase_service.Signup(data)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	if code != 200 {
		error_service.PublicError(c, response.Msg, 422, "password_confirmation", response.ErrorCode, "user")
		return
	}

	var email string = response.Email

	c.JSON(code, gin.H{"email": email})
}

func SetToken(c *gin.Context) {
	var data supabase_service.SetTokenBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	email, err := auth_service.ParseJWT(data.Token)

	if (err != nil) {
		error_service.PublicError(c, "Invalid token", 401, "token", "", "")
		return
	}

	auth_service.SetAuthCookie(c, data.Token, data.ExpiresIn)
	c.JSON(http.StatusOK, gin.H{"email": email})
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

	c.JSON(http.StatusOK, gin.H{"email": data.Email})
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
		error_service.PublicError(c, response.ErrorDescription, 401, "", "", "credentials")
		return
	}

	var email string = response.User.Email
	var token string = response.AccessToken
	var expiresIn int = response.ExpiresIn

	auth_service.SetAuthCookie(c, token, expiresIn)

	c.JSON(http.StatusOK, gin.H{"email": email})
}

func Logout(c *gin.Context) {
	code, err := supabase_service.Logout(c.GetString("token"))
	if (err != nil) {
		error_service.InternalError(c, err.Error())
		return
	}
	if code < 200 && code >= 300 {
		error_service.PublicError(c, "Could not logout", 401, "", "", "")
		return
	}

	auth_service.ExpireAuthCookie(c)

	c.JSON(http.StatusOK, gin.H{})
}
