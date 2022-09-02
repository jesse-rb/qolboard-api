package api

import (
	"fmt"
	"net/http"
	"qolboard-api/logger"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User schema
type User struct {
	gorm.Model
	Email string
	DisplayName string
	UUID string
}

// Input for post user request
type InputCreateUser struct {
	Email string `json:"email" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	PasswordConfirm string `json:"password_confirm" binding:"required"`
}

// Get all users
func GetUsers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var users []User
	db.Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// Create new user
func CreateUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	firebaseAuth := c.MustGet("firebaseAuth").(*auth.Client)

	// Validate input
	var input InputCreateUser
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Register user in firebase
	params := (&auth.UserToCreate{}).
        Email(input.Email).
        EmailVerified(false).
        Password(input.Password).
        DisplayName(input.DisplayName).
        Disabled(false)
	u, err := firebaseAuth.CreateUser(c, params)
	if err != nil {
		logger.LogError(prefix, fmt.Sprintf("error creating user: %v", err));
	}
	logger.LogInfo(prefix, fmt.Sprintf("Successfully created user: %v\n", u))

	// Store user in local DB
	user := User{Email: input.Email, DisplayName: input.DisplayName, UUID: u.UID}
	db.Create(&user)
	c.JSON(http.StatusOK, gin.H{"data": user})
}