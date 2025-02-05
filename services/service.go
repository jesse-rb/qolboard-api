package service

import (
	"crypto/rand"
	"encoding/base64"
	"qolboard-api/services/logging"
)

func GenerateCode(len uint) (string, error) {
	// Number of bytes needed for len base64 encoded chars
	lenBytes := (len*6 + 7) / 8

	// Init byte slice
	randomBytes := make([]byte, lenBytes)

	// Get random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		logging.LogError("generateCode", "Error generatiing code", err)
		return "", err
	}

	// Encode bytes to URL-safe Base64 string
	code := base64.RawURLEncoding.EncodeToString(randomBytes)

	return code, nil
}
