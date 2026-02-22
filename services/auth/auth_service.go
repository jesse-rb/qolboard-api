package auth_service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"qolboard-api/config"
	"qolboard-api/services/logging"

	"github.com/MicahParks/keyfunc/v3"
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
	supabaseHost := os.Getenv("SUPABASE_HOST")
	if supabaseHost == "" {
		logging.LogError("getJWKSURL", "Please set SUPABASE_HOST environment variable", "empty")
		panic(1)
	}
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", supabaseHost)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keyfunc, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc from jwks: %w", err)
	}

	// Parse token and verify signature and validate token issuer
	claims := &Claims{}
	withIssuer := jwt.WithIssuer(supabaseHost)
	_, err = jwt.ParseWithClaims(token, claims, keyfunc.Keyfunc, withIssuer)
	if err != nil {
		// Check if the token is valid
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	return claims, nil
}
