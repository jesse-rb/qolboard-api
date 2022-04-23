package config

import (
	"context"
	"encoding/base64"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var err error

func FirebaseAuth() *auth.Client {
	// Get google credentials
	googleCreds, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_CREDS")) // Decode base64 encoded google creds into byte array
	if err != nil {
		logError.Panic("Failed to decode google credentials")
	}
	
	// Use google credentials to to access firebase SDK
	opt := option.WithCredentialsJSON(googleCreds)
	// opt := option.WithCredentials(googleCredsStruct)
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