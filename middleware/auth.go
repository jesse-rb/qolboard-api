package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

// Authenticate request
func Auth(c *gin.Context) {
	firebaseAuth := c.MustGet("firebaseAuth").(*auth.Client)
	clientToken := c.GetHeader("Authorization")
	clientToken = strings.TrimSpace(strings.Replace(clientToken, "Bearer", "", 1))
	if clientToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id token not available"})
		c.Abort()
		return
	}
	
	// Check the token is valid
	token, err := firebaseAuth.VerifyIDToken(context.Background(), clientToken)
		if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}
	c.Set("UUID", token.UID)
	c.Next()
}