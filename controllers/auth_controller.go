package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"qolboard-api/config"
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
	"time"

	"github.com/gin-gonic/gin"
)

type registerBodyData struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *RESTHandler) Register(c *gin.Context) {
	var data registerBodyData

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
		error_service.PublicError(c, "a user with this email is already registered", http.StatusConflict, "email", data.Email, "user")
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
		thresholdUnix := time.Now().Unix() - int64(config.RateLimitRegister().Seconds())
		logging.LogDebug("DEBUG", "", map[string]any{"email verified": user.EmailVerificationCodeIAT})
		if user.EmailVerificationCodeIAT != nil && user.EmailVerificationCodeIAT.Unix() > thresholdUnix {
			// Not enough time has passed since the last verification email, abort, we want to be extra careful to never accidentally spam anyone, ever!
			error_service.PublicError(c, fmt.Sprintf("too soon, please wait %s between retries", config.RateLimitRegister().String()), http.StatusTooManyRequests, "email", data.Email, "user")
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

	resp := model.User{
		Email: data.Email,
	}

	response_service.SetJSON(c, gin.H{
		"data": resp,
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

	// If email has been verified, we can log the user in automatically
	jwt_token, err := auth_service.IssueJWT(user.Id)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	refresh_token, err := auth_service.IssueRefreshToken(tx, user.Id)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	auth_service.SetJWTCookie(c, jwt_token, int(config.TTLJWTToken().Seconds()))
	auth_service.SetRefreshTokenCookie(c, refresh_token, int(config.TTLRefreshToken().Seconds()))

	// Redirect
	appHost := os.Getenv("APP_HOST")
	locatoin := fmt.Sprintf("%s/canvas", appHost)
	c.Redirect(http.StatusFound, locatoin)
}

type requestOTPBodyData struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *RESTHandler) RequestOTP(c *gin.Context) {
	var data registerBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	tx, err := database_config.DB(nil)
	if err != nil || tx == nil {
		error_service.InternalError(c, err.Error())
		return
	}
	defer database.StandardDeferRollback(tx)

	// Get user by email
	user, err := user_model.GetByEmail(tx, data.Email)
	isErrNoRows := errors.Is(err, sql.ErrNoRows)
	if err != nil && !isErrNoRows {
		error_service.InternalError(c, err.Error())
		return
	}

	// Check if user is verified
	if isErrNoRows || user == nil {
		error_service.PublicError(c, "invalid user", http.StatusUnauthorized, "email", data.Email, "user")
		return
	}

	if user.VerifiedAt == nil {
		error_service.PublicError(c, "user is not verified", http.StatusBadRequest, "email", data.Email, "user")
		return
	}

	now := time.Now()
	if user.LoginOTPIAT != nil && user.LoginOTPIAT.Add(config.RateLimitRequestOTP()).After(now) {
		error_service.PublicError(c, fmt.Sprintf("too soon, please wait %s between retries", config.RateLimitRequestOTP().String()), http.StatusTooManyRequests, "email", data.Email, "user")
	}

	// Generate OTP
	code, err := service.GenerateCode(6)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	hashed := hashing.Sha256(code)

	// Save OTP
	user.LoginOTP = &hashed
	user.LoginOTPIAT = &now

	err = user.Update(tx, []string{"login_otp", "login_otp_iat"})
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// Send OTP to email
	email.SendOTPEmail(c, h.emailClient, user.Email, code)

	// If all went well, commit tx
	err = tx.Commit()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	resp := model.User{
		Email: data.Email,
	}

	response_service.SetJSON(c, map[string]any{
		"data": resp,
	})
}

type loginBodyData struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

func (h *RESTHandler) Login(c *gin.Context) {
	var data loginBodyData

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	// Start tx
	tx, err := database_config.DB(nil)
	defer database.StandardDeferRollback(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
	}

	// Get user by email
	user, err := user_model.GetByEmail(tx, data.Email)
	isErrNoRows := errors.Is(err, sql.ErrNoRows)
	if err != nil && !isErrNoRows {
		error_service.InternalError(c, err.Error())
	}

	// Check if valid user
	if isErrNoRows || user == nil {
		error_service.PublicError(c, "invalid user", http.StatusUnauthorized, "email", data.Email, "user")
		return
	}

	// Hash OTP
	hashed := hashing.Sha256(data.OTP)

	// Verify hashed OTP match
	now := time.Now()
	if !(user.LoginOTP != nil && user.LoginOTPIAT != nil && user.LoginOTPIAT.Add(config.TTLLoginOTP()).After(now) && *user.LoginOTP == hashed) {
		error_service.PublicError(c, "invalid opt", http.StatusUnauthorized, "otp", data.OTP, "user")
		return
	}

	// Expire OTP
	user.LoginOTP = nil
	user.LoginOTPIAT = nil
	err = user.Update(tx, []string{"login_otp", "login_otp_iat"})
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// Issue JWT token
	token, err := auth_service.IssueJWT(user.Id)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	refresh_token, err := auth_service.IssueRefreshToken(tx, user.Id)
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

	auth_service.SetRefreshTokenCookie(c, refresh_token, int(config.TTLRefreshToken().Seconds()))

	// Set token
	auth_service.SetJWTCookie(c, token, int(config.TTLJWTToken().Seconds()))

	// Redirect to
	appHost := os.Getenv("APP_HOST")
	locatoin := fmt.Sprintf("%s/canvas", appHost)
	c.Redirect(http.StatusFound, locatoin)
}

func (h *RESTHandler) Logout(c *gin.Context) {
	// Force expire refresh token
	userID := auth_service.Auth(c)
	tx, err := database_config.DB(nil)
	if err != nil {
		error_service.InternalError(c, err.Error())
	} else {
		auth_service.ForceExpireRefreshToken(c, tx, userID)
	}

	// Expire cookies
	auth_service.ExpireJWTCookie(c)
	auth_service.ExpireRefreshTokenCookie(c)
}
