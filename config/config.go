package config

import (
	"os"
	"time"
)

func IsDev() bool {
	return os.Getenv("ENV") == "dev"
}

func LoginOTPTTL() time.Duration {
	return 5 * time.Minute
}

func JWTTTL() time.Duration {
	return time.Minute * 15
}
