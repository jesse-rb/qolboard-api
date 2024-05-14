package test_controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	var data string = "Hello world"
	c.JSON(http.StatusOK, gin.H{"data": data})
}
