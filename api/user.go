package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User schema
type User struct {
	gorm.Model
	Email string
	Name string
}

// Input for post user request
type InputPostUser struct {
	Email string `json:"email" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// Get all users
func GetUsers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var users []User
	db.Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// Create new user
func PostUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	// Validate input
	var input InputPostUser
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Create artist
	user := User{Email: input.Email, Name: input.Name}
	db.Create(&user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}