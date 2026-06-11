package service

import (
	"net/http"
	"os"
	"qolboard-api/config"

	"github.com/gin-gonic/gin"
)

func SetCookie(c *gin.Context, name string, value string, maxAge int, path string) {
	var (
		domain string = os.Getenv("APP_DOMAIN")
		secure bool   = true
		isDev  bool   = config.IsDev()
	)

	if isDev {
		secure = false
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie(name, value, maxAge, path, domain, secure, true) // Expire jwt cookie
}
