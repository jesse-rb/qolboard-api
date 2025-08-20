package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"qolboard-api/services/logging"
)

// Generates a random code
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

// Returns an empty map[string]any if an error occured
// This may not be performant enough one day
func ToMapStringAny(probablyAStruct any) map[string]any {
	toReturn := make(map[string]any)

	// Marshal into JSON byte array
	buf, err := json.Marshal(probablyAStruct)
	if err != nil {
		logging.LogError("[service]", "Failed json marhsal", err)
		return toReturn
	}

	// Unmarshal into map[string]any
	err = json.Unmarshal(buf, &toReturn)
	if err != nil {
		logging.LogError("[service]", "Failed json unmarshal", err)
		return toReturn
	}

	return toReturn
}

func PrettyJson(v any) string {
	pretty := ""
	out, err := json.MarshalIndent(v, "", "\t")
	if err == nil {
		pretty = string(out)
	}
	return pretty
}
