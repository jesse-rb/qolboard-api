package canvas_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	database_config "qolboard-api/config/database"
	canvas_model "qolboard-api/models/canvas"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger slogger.Logger = *slogger.New(os.Stdout, slogger.ANSIGreen, "canvas_controller", log.Lshortfile+log.Ldate)
var errorLogger slogger.Logger = *slogger.New(os.Stderr, slogger.ANSIRed, "canvas_controller", log.Lshortfile+log.Ldate)

func Index(c *gin.Context) {
	db := database_config.GetDatabase();

	email, _ := c.Get("email")

	var Canvases []*canvas_model.Canvas

	db.Connection.Where("user_email = ?", email).Find(&Canvases)

	c.JSON(http.StatusOK, Canvases)
}

func Get(c *gin.Context) {
	db := database_config.GetDatabase();

	email, _ := c.Get("email")

	var id string = c.Param("id")

	var Canvas canvas_model.Canvas

	db.Connection.Where("user_email = ?", email).First(&Canvas, id)

	c.JSON(http.StatusOK, Canvas)
}

func Save(c *gin.Context) {
	db := database_config.GetDatabase();

	email := c.GetString("email")

	var canvasData canvas_model.CanvasData
	if err := c.ShouldBindJSON(&canvasData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return;
	}

	canvasDataJson, err := json.Marshal(canvasData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var canvas canvas_model.Canvas = canvas_model.Canvas{UserEmail: email, CanvasData: canvasDataJson}

	result := db.Connection.Create(&canvas)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving canvas data"})

		errorLogger.Log("Save", "Error saving canvas data", result)
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprintf("Successfully saved canvas with id: %v", canvas.ID),
		"canvas": canvas,
	})
}