package supabase_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	model "qolboard-api/models"
	"qolboard-api/services/logging"
)

type RegisterBodyData struct {
	Email                string `json:"email" binding:"required,email"`
	Password             string `json:"password" binding:"required"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required"`
}

type LoginBodyData struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SetTokenBodyData struct {
	Token     string `json:"token" binding:"required"`
	ExpiresIn int    `json:"expires_in" binding:"required"`
}

type ResendEmailVerificationBodyData struct {
	Email string `json:"email" binding:"email,required"`
}

type ResendBodyData struct {
	Type  string `json:"type" binding:"required"`
	Email string `json:"email" binding:"email,required"`
}

type SupabaseResponse struct {
	ErrorCode string `json:"error_code"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
}

type SupabaseRegisterResponse struct {
	Email string `json:"email"`
	Uuid  string `json:"id"`
	SupabaseResponse
}

type SupabaseLoginResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"`
	User        model.User `json:"user"`
	SupabaseResponse
}

type SupabaseLogoutResponse struct {
	SupabaseResponse
}

func Signup(data RegisterBodyData) (code int, supabaseRegisterResponse *SupabaseRegisterResponse, err error) {
	requestBody, _ := json.Marshal(data)
	code, response, err := supabase(http.MethodPost, "signup", requestBody, "")
	if err != nil {
		return code, nil, err
	}

	logging.LogDebug("supabase_service::Signup", "response", string(response))
	logging.LogInfo("Signup", "received supabase signup response with code", code)

	var supabaseResponse SupabaseRegisterResponse
	err = json.Unmarshal(response, &supabaseResponse)
	if err != nil {
		return code, nil, err
	}

	return code, &supabaseResponse, err
}

func Login(data LoginBodyData) (code int, supabaseLoginResponse *SupabaseLoginResponse, err error) {
	requestBody, _ := json.Marshal(data)

	code, response, err := supabase(http.MethodPost, "token?grant_type=password", requestBody, "")
	if err != nil {
		return code, nil, err
	}

	logging.LogInfo("Login", "received supabase login with code", code)

	var supabaseResponse SupabaseLoginResponse
	err = json.Unmarshal(response, &supabaseResponse)
	if err != nil {
		return code, nil, err
	}

	return code, &supabaseResponse, nil
}

func Logout(token string) (code int, err error) {
	code, _, err = supabase(http.MethodPost, "logout", nil, token)
	if err != nil {
		return code, err
	}

	logging.LogInfo("Logout", "received supabase logout with code", code)

	return code, err
}

func ForgotPassword() (err error) {
	_, _, err = supabase(http.MethodPost, "recover", nil, "")
	return err
}

func ResendEmailVerification(email string) (int, error) {
	var data ResendBodyData = ResendBodyData{
		Type:  "signup",
		Email: email,
	}
	requestBody, _ := json.Marshal(data)

	code, _, err := supabase(http.MethodPost, "resend", requestBody, "")
	if err != nil {
		return code, err
	}

	return code, err
}

func supabase(method string, path string, bodyData []byte, token string) (code int, responseBodyBytes []byte, err error) {
	var host string = os.Getenv("SUPABASE_HOST")
	var url string = fmt.Sprintf("%s/%s", host, path)
	var apiKey string = os.Getenv("SUPABASE_ANON_KEY")

	request, err := http.NewRequest(method, url, bytes.NewBuffer(bodyData))
	if err != nil {
		logging.LogError("supabase", "Failed initiating supabase request", err.Error())
		return 0, nil, err
	}

	if token != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	request.Header.Set("apikey", apiKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logging.LogError("supabase", "Failed sending supabase request", err.Error())
		return response.StatusCode, nil, err
	}

	defer response.Body.Close()

	responseBodyBytes, err = io.ReadAll(response.Body)
	if err != nil {
		logging.LogError("supabase", "Failed reading supabase response body", err.Error())
		return response.StatusCode, nil, err
	}

	return response.StatusCode, responseBodyBytes, err
}
