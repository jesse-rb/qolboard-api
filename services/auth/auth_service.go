package auth_service

import (
	"fmt"
	"os"
	"qolboard-api/config"
	model "qolboard-api/models"
	service "qolboard-api/services"
	"qolboard-api/services/hashing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
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

func SetJWTCookie(c *gin.Context, token string, expiresIn int) {
	service.SetCookie(c, "qolboard_jwt", token, expiresIn, "/")
}

func ExpireJWTCookie(c *gin.Context) {
	service.SetCookie(c, "qolboard_jwt", "", 0, "/") // Expire jwt cookie
}

func SetRefreshTokenCookie(c *gin.Context, token string, expiresIn int) {
	service.SetCookie(c, "qolboard_refresh_token", token, expiresIn, "/")
}

func ExpireRefreshTokenCookie(c *gin.Context) {
	service.SetCookie(c, "qolboard_refresh_token", "", 0, "/") // Expire refresh token cookie
}

func GetJWTCookie(c *gin.Context) (string, error) {
	token, err := c.Cookie("qolboard_jwt")
	if err != nil {
		return "", fmt.Errorf("failed to get jwt cookie: %w", err)
	}

	return token, nil
}

func GetRefreshTokenCookie(c *gin.Context) (string, error) {
	token, err := c.Cookie("qolboard_refresh_token")
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token cookie: %w", err)
	}

	return token, nil
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
		claims, _ := verifiedToken.Claims.(*Claims)
		return claims, fmt.Errorf("error verifying token: %w", err)
	}

	// Ensure token is valid, and we can get claims
	claims, ok := verifiedToken.Claims.(*Claims)
	if !ok || !verifiedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func IssueJWT(userID string) (string, error) {
	iss := os.Getenv("API_HOST")
	now := time.Now()
	iat := jwt.NewNumericDate(now)
	exp := jwt.NewNumericDate(now.Add(config.TTLJWTToken())) // JWT expires in 15 minutes from now
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss,
			Subject:   userID,
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

func IssueRefreshToken(tx *sqlx.Tx, userID string) (string, error) {
	token, err := service.GenerateCode(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	hashed := hashing.Sha256(token)

	urt := model.UserRefreshToken{
		UserID:       userID,
		RefreshToken: hashed,
	}

	err = urt.Create(tx)
	if err != nil {
		return "", fmt.Errorf("error creating user refresh token: %w", err)
	}

	return token, nil
}

func ForceExpireRefreshToken(c *gin.Context, tx *sqlx.Tx, userID string) error {
	token, err := GetRefreshTokenCookie(c)
	if err != nil {
		return fmt.Errorf("failed to get refresh token cookie: %w", err)
	}

	urt := model.UserRefreshToken{
		UserID:       userID,
		RefreshToken: token,
	}

	err = urt.DeleteByRefreshToken(tx)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func ValidateRefreshToken(c *gin.Context, tx *sqlx.Tx, userID string) error {
	token, err := GetRefreshTokenCookie(c)
	if err != nil {
		return fmt.Errorf("failed to get refresh token cookie: %w", err)
	}

	urt := model.UserRefreshToken{
		UserID:       userID,
		RefreshToken: token,
	}

	err = urt.FindByRefreshToken(tx)
	if err != nil {
		return fmt.Errorf("failed to find refresh token: %w", err)
	}

	now := time.Now()

	if urt.CreatedAt.Add(config.TTLRefreshToken()).Before(now) {
		return fmt.Errorf("refresh token expired")
	}

	return nil
}
