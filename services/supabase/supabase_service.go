package supabase_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "supabase_service", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "supabase_service", log.Lshortfile+log.Ldate);

type RegisterBodyData struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required"`
}

type LoginBodyData struct {
	Email  string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SetTokenBodyData struct {
	Token string `json:"token" binding:"required"`
	ExpiresIn int `json:"expires_in" binding:"required"`
}

type ResendEmailVerificationBodyData struct {
	Email string `json:"email" binding:"email,required"`
}

type ResendBodyData struct {
	Type string `json:"type" binding:"required"`
	Email string `json:"email" binding:"email,required"`
}

type User struct {
	Email string `json:"email"`
}

type SupabaseRegisterResponse struct {
	Email string `json:"email"`
	ErrorCode string `json:"error_code"`
	Msg string `json:"msg"`
}

type SupabaseLoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	User User `json:"user"`
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type SupabaseLogoutResponse struct {
	ErrorCode string `json:"error_code"`
	Msg string `json:"msg"`
}

func Signup(data RegisterBodyData) (code int, supabaseRegisterResponse *SupabaseRegisterResponse, err error) {
	var requestBody, _ = json.Marshal(data)
	code, response, err := supabase(http.MethodPost, "signup", requestBody, "")

	if err != nil {
		return code, nil, err
	}

	infoLogger.Log("Signup", "received supabase signup response with code", code)

	var supabaseResponse SupabaseRegisterResponse
	err = json.Unmarshal(response, &supabaseResponse)
	if err != nil {
		return code, nil, err
	}
	
	return code, &supabaseResponse, err
}

func Login(data LoginBodyData) (code int, supabaseLoginResponse *SupabaseLoginResponse, err error) {
	var requestBody, _ = json.Marshal(data)

	code, response, err := supabase(http.MethodPost, "token?grant_type=password", requestBody, "")
	if err != nil {
		return code, nil, err
	}

	infoLogger.Log("Login", "received supabase login with code", code)

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

	infoLogger.Log("Logout", "received supabase logout with code", code)

	return code, err
}

func ForgotPassword() (err error) {
	_, _, err = supabase(http.MethodPost, "recover", nil, "")
	return err
}

func ResendEmailVerification(email string) (int, error) {
	var data ResendBodyData = ResendBodyData{
		Type: "signup",
		Email: email,
	}
	var requestBody, _ = json.Marshal(data)

	code, _, err := supabase(http.MethodPost, "resend", requestBody, "")
	if err != nil {
		return code, err
	}

	return code, err
}

func supabase(method string, path string, bodyData []byte, token string) (code int, responseBodyBytes []byte, err error) {
	var host string = os.Getenv("SUPABASE_HOST")
	var url string = fmt.Sprintf("%s/%s", host, path);
	var apiKey string = os.Getenv("SUPABASE_ANON_KEY")

	request, err := http.NewRequest(method, url, bytes.NewBuffer(bodyData))
	if err != nil {
		errorLogger.Log("supabase", "Failed initiating supabase request", err);
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
        errorLogger.Log("supabase", "Failed sending supabase request", err);
		return response.StatusCode, nil, err
    }
	
	defer response.Body.Close()

	responseBodyBytes, err = io.ReadAll(response.Body)
	if err != nil {
		errorLogger.Log("supabase", "Failed reading supabase response body", err)
		return response.StatusCode, nil, err
	}

	return response.StatusCode, responseBodyBytes, err
}