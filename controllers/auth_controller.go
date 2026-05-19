package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	user_model "qolboard-api/models/user"
	service "qolboard-api/services"
	auth_service "qolboard-api/services/auth"
	"qolboard-api/services/database"
	"qolboard-api/services/email"
	error_service "qolboard-api/services/error"
	"qolboard-api/services/hashing"
	"qolboard-api/services/logging"
	response_service "qolboard-api/services/response"
	supabase_service "qolboard-api/services/supabase"
	"time"

	"github.com/gin-gonic/gin"
)

type RegisterBodyData struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *RESTHandler) Register(c *gin.Context) {
	var data RegisterBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	tx, err := database_config.DB(nil)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	defer database.StandardDeferRollback(tx)

	// Check if user already exists
	user, err := user_model.GetByEmail(tx, data.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		error_service.InternalError(c, err.Error())
		return
	}

	if user != nil && user.VerifiedAt != nil {
		error_service.PublicError(c, "a user with this email is already registered", http.StatusBadRequest, "email", data.Email, "user")
		return
	}

	token, err := service.GenerateCode(256)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// Hash the token
	hashed := hashing.Sha256(token)
	iat := time.Now()

	if user != nil {
		thresholdUnix := time.Now().Unix() - 60
		logging.LogDebug("DEBUG", "", map[string]any{"email verified": user.EmailVerificationCodeIAT})
		if user.EmailVerificationCodeIAT != nil && user.EmailVerificationCodeIAT.Unix() > thresholdUnix {
			// Not enough time has passed since the last verification email, abort, we want to be extra careful to never accidentally spam anyone, ever!
			error_service.PublicError(c, "too soon, please wait 1 minute between retries", http.StatusTooManyRequests, "email", data.Email, "user")
			return
		}

		user.EmailVerificationCode = &hashed
		user.EmailVerificationCodeIAT = &iat

		err = user.Update(tx, []string{"email_verification_code", "email_verification_code_iat"})
		if err != nil {
			error_service.InternalError(c, err.Error())
			return
		}
	} else {
		newUser := &model.User{
			Email:                    data.Email,
			EmailVerificationCode:    &hashed,
			EmailVerificationCodeIAT: &iat,
		}

		err = newUser.Create(tx)
		if err != nil {
			error_service.InternalError(c, err.Error())
			return
		}
	}

	err = email.SendVerificationEmail(c, h.emailClient, data.Email, token)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// If all went well, commit tx
	err = tx.Commit()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": "",
	})
}

type verifyEmailParams struct {
	Token string `form:"token" binding:"required"`
}

func (h *RESTHandler) VerifyEmail(c *gin.Context) {
	var params = verifyEmailParams{
		Token: "",
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		error_service.ValidationError(c, err)
		return
	}

	hashed := hashing.Sha256(params.Token)

	tx, err := database_config.DB(nil)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	defer database.StandardDeferRollback(tx)

	user, err := user_model.GetByEmailVerificationCode(tx, hashed)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		error_service.InternalError(c, err.Error())
		return
	}

	if err != nil || user == nil {
		error_service.PublicError(c, "invalid email verification code", http.StatusUnauthorized, "email_verification_code", "", "value")
		return
	}

	verifiedAt := time.Now()
	user.EmailVerificationCode = nil
	user.EmailVerificationCodeIAT = nil
	user.VerifiedAt = &verifiedAt
	err = user.Update(tx, []string{"verified_at", "email_verification_code", "email_verification_code_iat"})
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// If all went well, commit tx
	err = tx.Commit()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": "",
	})
}

func (h *RESTHandler) SetToken(c *gin.Context) {
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

func (h *RESTHandler) ResendVerificationEmail(c *gin.Context) {
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

func (h *RESTHandler) Login(c *gin.Context) {
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

	// Redirect to /user
	appHost := os.Getenv("APP_HOST")
	locatoin := fmt.Sprintf("%s/user", appHost)
	c.Redirect(http.StatusFound, locatoin)
}

func (h *RESTHandler) Logout(c *gin.Context) {
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
