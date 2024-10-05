package auth_service

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "auth_service", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "auth_service", log.Lshortfile+log.Ldate);

var domain string = os.Getenv("APP_DOMAIN")
var secure bool = true;
var isDev bool = os.Getenv("GIN_MODE") == "dev"

func init() {
	
}

func SetAuthCookie(c *gin.Context, token string, expiresIn int) {
	if isDev {
		secure = false;
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", token, expiresIn, "/", domain, secure, true)
}

func ExpireAuthCookie(c *gin.Context) {
	if isDev {
		secure = false;
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", "", 0, "/", domain, secure, true) // Expire jwt cookie
}