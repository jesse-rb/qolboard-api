package auth_service

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "auth_service", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "auth_service", log.Lshortfile+log.Ldate);

var domain string = os.Getenv("APP_DOMAIN")
var secure bool = true;
var isDev bool = os.Getenv("GIN_MODE") == "dev"

func init() {
	
}

type claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func ParseJWT(token string) (email string, err error) {
	var secret string = os.Getenv("SUPABASE_JWT_SECRET")
	if (secret == "") {
		errorLogger.Log("parseJWT", "Please set SUPABASE_JWT_SECRET environment variable", "empty")
		panic(1)
	}

	// Parse token and validate signature
	t, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("error unexpected signing method: %v", t.Header["alg"])
        } 
        return []byte(secret), nil
	})

	// Check if the token is valid
	if err != nil {
		return "", fmt.Errorf("error validating tokenL %v", err)
	} else if claims, ok := t.Claims.(*claims); ok {
		return claims.Email, nil
	}

	return "", fmt.Errorf("error parsing token: %v", err)
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