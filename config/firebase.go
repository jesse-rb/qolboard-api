package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var err error

func FirebaseAuth() *auth.Client {
	// Get google credentials
	googleCreds, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_CREDS")) // Decode base64 encoded google creds
	if err != nil {
		logError.Panic("Failed to decode google credentials")
	}
	googleCredsStruct := &google.Credentials{}
	err = json.Unmarshal([]byte(googleCreds), &googleCredsStruct) // Unmarhsal google creds
	if err != nil {
		logError.Panic("Failed to unmarshal google credentials")
	}

	// Use google credentials to to access firebase SDK
	opt := option.WithCredentials(googleCredsStruct)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logError.Panic("Failed to load firebase app")
	}
	
	// Init firebase auth
	auth, err := app.Auth(context.Background())
	if err != nil {
		logError.Panic("Failed to load firebase app auth")
	}
	
	return auth
}