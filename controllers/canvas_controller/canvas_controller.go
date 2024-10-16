package canvas_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	database_config "qolboard-api/config/database"
	canvas_model "qolboard-api/models/canvas"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger slogger.Logger = *slogger.New(os.Stdout, slogger.ANSIGreen, "canvas_controller", log.Lshortfile+log.Ldate)
var errorLogger slogger.Logger = *slogger.New(os.Stderr, slogger.ANSIRed, "canvas_controller", log.Lshortfile+log.Ldate)

func Index(c *gin.Context) {
	db := database_config.GetDatabase();

	email := c.GetString("email");

	var canvases []*canvas_model.Canvas;

	db.Connection.Scopes(canvas_model.BelongsToUser(email)).Find(&canvases);

	response_service.SetJSON(c, gin.H{
		"data": canvases,
	})
}

func Get(c *gin.Context) {
	db := database_config.GetDatabase();

	email := c.GetString("email");

	var id string = c.Param("id");

	var canvas canvas_model.Canvas;

	db.Connection.Scopes(canvas_model.BelongsToUser(email)).First(&canvas, id);

	response_service.SetJSON(c, canvas);
}

func Save(c *gin.Context) {
	db := database_config.GetDatabase();

	email := c.GetString("email");

	var paramId string = c.Param("id");
	var id uint64 = 0;
	var err error = nil;
	if paramId != "" {
		id, err = strconv.ParseUint(paramId, 10, 64);
		if err != nil {
			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
			return
		}
	}

	var canvasData canvas_model.CanvasData
	if err := c.ShouldBindJSON(&canvasData); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	canvasDataJson, err := json.Marshal(canvasData)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return;
	}
	
	var canvas canvas_model.Canvas = canvas_model.Canvas{UserEmail: email, CanvasData: canvasDataJson};

	if id > 0 {
		// Update
		canvas.ID = id;
	}

	result := db.Connection.
		Scopes(canvas_model.BelongsToUser(email)).
		Save(&canvas);

	if result.Error != nil {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"msg": fmt.Sprintf("Successfully saved canvas with id: %v", canvas.ID),
		"canvas": canvas,
	})
}

func Delete(c *gin.Context) {
	db := database_config.GetDatabase();

	email := c.GetString("email");

	var paramId string = c.Param("id");
	var id uint64 = 0;
	var err error = nil;
	if paramId != "" {
		id, err = strconv.ParseUint(paramId, 10, 64);
		if err != nil {
			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
			return
		}
	}

	var canvas canvas_model.Canvas;

	canvas.ID = id;

	db.Connection.
		Scopes(canvas_model.BelongsToUser(email)).
		First(&canvas, id);

	result := db.Connection.
		Scopes(canvas_model.BelongsToUser(email)).
		Delete(&canvas, id);

	if (result.Error != nil) {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": fmt.Sprintf("Successfully saved canvas with id %v", canvas.ID),
		"data": canvas,
	})
}