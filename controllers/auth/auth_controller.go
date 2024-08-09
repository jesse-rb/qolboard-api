package auth_controller

import (
	"log"
	"net/http"
	"os"
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
		// errorLogger.Log("Register", "Failed supabase signup", err.Error())
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

	var domain string = os.Getenv("APP_DOMAIN")
	var secure bool = true

	var isDev bool = os.Getenv("GIN_MODE") == "dev"
	if isDev {
		secure = false;
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", token, expiresIn, "/", domain, secure, true)

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

	var domain string = os.Getenv("APP_DOMAIN")
	var secure bool = true

	var isDev bool = os.Getenv("GIN_MODE") == "dev"
	if isDev {
		secure = false;
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", "", 0, "/", domain, secure, true) // Expire jwt cookie

	c.JSON(http.StatusOK, gin.H{})
}
