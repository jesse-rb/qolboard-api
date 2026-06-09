package config

import (
	"os"
	"time"
)

func IsDev() bool {
	return os.Getenv("ENV") == "dev"
}

func TTLJWTToken() time.Duration {
	return 15 * time.Minute
}

func TTLLoginOTP() time.Duration {
	return 5 * time.Minute
}

func TTLEmailVerificationToken() time.Duration {
	return 24 * time.Hour
}

func RateLimitRegister() time.Duration {
	return 24 * time.Hour
}

func RateLimitRequestOTP() time.Duration {
	return 15 * time.Minute
}
