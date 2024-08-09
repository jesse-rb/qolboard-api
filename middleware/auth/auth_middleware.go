package auth_middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "auth_middleware", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "auth_middleware", log.Lshortfile+log.Ldate);

type claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func parseJWT(token string) (email string, err error) {
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

// Authenticate middleware
func Run(c *gin.Context) {
	token, err := c.Cookie("qolboard_jwt")
	
	if (err != nil) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if (token == "") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	email, err := parseJWT(token)

	if (err != nil) {
		infoLogger.Log("AuthMiddleware", "Error parsing token", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	infoLogger.Log("AuthMiddleware", "Received request from", email)

	c.Set("email", email)
	c.Set("token", token)
	c.Next()
}
