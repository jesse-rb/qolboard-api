package config

import "os"

func IsDev() bool {
	return os.Getenv("ENV") == "dev"
}
