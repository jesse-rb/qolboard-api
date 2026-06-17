package auth_middleware

import (
	"errors"
	"net/http"
	database_config "qolboard-api/config/database"
	auth_service "qolboard-api/services/auth"
	"qolboard-api/services/database"
	error_service "qolboard-api/services/error"
	"qolboard-api/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Authenticate middleware
func Run(c *gin.Context) {
	token, err := auth_service.GetJWTCookie(c)
	if err != nil {
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	if token == "" {
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	claims, err := auth_service.ParseJWT(token)
	if err != nil {
		logging.LogDebug("AuthMiddleware", "token not valid", claims)
		if errors.Is(err, jwt.ErrTokenExpired) {
			// If the token is expired but cryptographically valid, see if we can refresh the token
			logging.LogDebug("AuthMiddleware", "jwt token is expired, attempting refresh", nil)

			tx, err := database_config.DB(nil)
			if err != nil {
				error_service.InternalError(c, err.Error())
				c.Abort()
				return
			}
			database.StandardDeferRollback(tx)

			userID := claims.Subject

			familyID, err := auth_service.ValidateRefreshToken(c, tx, userID)
			if err == nil {
				// Refresh token is valid, so issue new JWT and refresh token
				auth_service.ForceExpireRefreshToken(c, tx, userID)
				auth_service.IssueJWT(userID)
				auth_service.IssueRefreshToken(tx, userID, familyID)
			}
		}

		logging.LogInfo("AuthMiddleware", "Error parsing token", err)
		error_service.PublicError(c, "Unauthorized", http.StatusUnauthorized, "", "", "user")
		c.Abort()
		return
	}

	logging.LogInfo("AuthMiddleware", "Received request from", claims.Email)

	c.Set("claims", claims)
	c.Set("token", token)
	c.Next()
}
