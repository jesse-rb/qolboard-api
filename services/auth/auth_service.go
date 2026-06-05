package auth_service

import (
	"fmt"
	"net/http"
	"os"
	"qolboard-api/config"
	model "qolboard-api/models"
	"time"

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

// Gets the authenticated user's claims from the JWT
// panics if unsuccessful
func GetClaims(c *gin.Context) *Claims {
	claimsAny, exists := c.Get("claims")
	if !exists {
		panic("No claims set")
	}

	if claims, ok := claimsAny.(*Claims); ok {
		return claims
	} else {
		panic("Unexpected claims structure")
	}
}

// Gets the authenticated user's uuid based on the JWT claims
// panics if unsuccessful
func Auth(c *gin.Context) string {
	claims := GetClaims(c)
	return claims.Subject
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
	iss := os.Getenv("API_HOST")
	keyfunc := func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpcected signing method: %v", token.Header["alg"])
		}

		secret := os.Getenv("ACCESS_TOKEN_SECRET")
		return []byte(secret), nil
	}

	// Parse token and verify signature and validate token issuer
	claims := &Claims{}
	withIssuer := jwt.WithIssuer(iss)
	verifiedToken, err := jwt.ParseWithClaims(token, claims, keyfunc, withIssuer)
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	// Ensure token is valid, and we can get claims
	claims, ok := verifiedToken.Claims.(*Claims)
	if !ok || !verifiedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func IssueJWT(user model.User) (string, error) {
	iss := os.Getenv("API_HOST")
	now := time.Now()
	iat := jwt.NewNumericDate(now)
	exp := jwt.NewNumericDate(now.Add(config.JWTTTL())) // JWT expires in 15 minutes from now
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss,
			Subject:   user.Uuid,
			IssuedAt:  iat,
			ExpiresAt: exp,
		},
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	token, err := unsigned.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return token, nil
}
