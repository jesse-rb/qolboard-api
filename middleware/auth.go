package middleware

import (
	"github.com/gin-gonic/gin"
)

// Authenticate request
func Auth(c *gin.Context) {
	
	// c.Set("UUID", token.UID)
	c.Next()
}