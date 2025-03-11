package auth_service

import (
	"fmt"
	"net/http"
	"os"
	"qolboard-api/config"
	"qolboard-api/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	domain string = os.Getenv("APP_DOMAIN")
	secure bool   = true
	isDev  bool   = config.IsDev()
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GetClaims(c *gin.Context) *Claims {
	claims, exists := c.Get("claims")
	if !exists {
		panic("No claims set")
	}

	return claims.(*Claims)
}

func SetAuthCookie(c *gin.Context, token string, expiresIn int) {
	if isDev {
		secure = false
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", token, expiresIn, "/", domain, secure, true)
}

func ExpireAuthCookie(c *gin.Context) {
	if isDev {
		secure = false
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("qolboard_jwt", "", 0, "/", domain, secure, true) // Expire jwt cookie
}

func ParseJWT(token string) (*Claims, error) {
	var secret string = os.Getenv("SUPABASE_JWT_SECRET")
	if secret == "" {
		logging.LogError("parseJWT", "Please set SUPABASE_JWT_SECRET environment variable", "empty")
		panic(1)
	}

	// Parse token and validate signature
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	// Check if the token is valid
	if err != nil {
		return nil, fmt.Errorf("error validating tokenL %v", err)
	} else if claims, ok := t.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("error parsing token: %v", err)
}
