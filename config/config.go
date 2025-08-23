package config

import "os"

func IsDev() bool {
	return os.Getenv("GIN_MODE") == "dev"
}
