package canvas_controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	"qolboard-api/services/logging"
	response_service "qolboard-api/services/response"
	websocket_service "qolboard-api/services/websocket"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Index(c *gin.Context) {
	db := database_config.GetDatabase()

	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	var canvases []*model.Canvas

	db.Connection.Scopes(model.CanvasBelongsToUser(userUuid)).Find(&canvases)

	response_service.SetJSON(c, gin.H{
		"data": canvases,
	})
}

func Get(c *gin.Context) {
	db := database_config.GetDatabase()

	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	var paramId string = c.Param("canvas_id")
	id, err := strconv.ParseUint(paramId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Canvas id must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas")
		return
	}

	var canvas model.Canvas = model.Canvas{}
	canvas.ID = id

	result := db.Connection.
		Joins("LEFT JOIN canvas_shared_accesses ON canvas_shared_accesses.canvas_id = canvas.id AND canvas_shared_accesses.user_uuid = ? AND canvas_shared_accesses.deleted_at IS NULL", userUuid).
		Where(db.Connection.Scopes(model.CanvasBelongsToUser(userUuid))).
		Or(db.Connection.Where("canvas_shared_accesses.user_uuid = ?", userUuid)).
		Preload("User").
		Preload("CanvasSharedAccess").
		Preload("CanvasSharedAccess.User").
		Preload("CanvasSharedInvitation").
		First(&canvas)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			error_service.PublicError(c, "Canvas not found", http.StatusNotFound, "id", paramId, "canvas")
			return
		}
		error_service.InternalError(c, result.Error.Error())
	}

	if canvas.CanvasSharedInvitation != nil {
		for i, csi := range canvas.CanvasSharedInvitation {
			canvas.CanvasSharedInvitation[i] = csi.Response()
		}
	}

	response_service.SetJSON(c, canvas)
}

func Save(c *gin.Context) {
	db := database_config.GetDatabase()

	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	var paramId string = c.Param("canvas_id")
	var id uint64 = 0
	var err error = nil
	if paramId != "" {
		id, err = strconv.ParseUint(paramId, 10, 64)
		if err != nil {
			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
			return
		}
	}

	logging.LogDebug("Save", "id", id)

	var canvasData model.CanvasData
	if err := c.ShouldBindJSON(&canvasData); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	canvasDataJson, err := json.Marshal(canvasData)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	var canvas model.Canvas = model.Canvas{}
	var result *gorm.DB
	if id > 0 {
		// Update
		canvas.ID = id

		result = db.Connection.
			Joins("LEFT JOIN canvas_shared_accesses ON canvas_shared_accesses.canvas_id = canvas.id AND canvas_shared_accesses.user_uuid = ? AND canvas_shared_accesses.deleted_at IS NULL", userUuid).
			Where(db.Connection.Scopes(model.CanvasBelongsToUser(userUuid))).
			Or(db.Connection.Where("canvas_shared_accesses.user_uuid = ?", userUuid)).
			First(&canvas)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				error_service.PublicError(c, "Canvas not found", http.StatusNotFound, "id", paramId, "canvas")
				return
			}

			error_service.InternalError(c, result.Error.Error())
			return
		}

		canvas.CanvasData = canvasDataJson
		result = db.Connection.Save(&canvas)
	} else {
		canvas.UserUuid = userUuid
		canvas.CanvasData = canvasDataJson

		result = db.Connection.
			Where(db.Connection.Scopes(model.CanvasBelongsToUser(userUuid))).
			Save(&canvas)
	}

	if result.Error != nil {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"msg":    fmt.Sprintf("Successfully saved canvas with id: %v", canvas.ID),
		"canvas": canvas,
	})
}

func Delete(c *gin.Context) {
	db := database_config.GetDatabase()

	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	var paramId string = c.Param("canvas_id")
	var id uint64 = 0
	var err error = nil
	if paramId != "" {
		id, err = strconv.ParseUint(paramId, 10, 64)
		if err != nil {
			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
			return
		}
	}

	var canvas model.Canvas

	canvas.ID = id

	db.Connection.
		Scopes(model.CanvasBelongsToUser(userUuid)).
		First(&canvas, id)

	result := db.Connection.
		Scopes(model.CanvasBelongsToUser(userUuid)).
		Delete(&canvas, id)

	if result.Error != nil {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": fmt.Sprintf("Successfully deleted canvas shared invitation with id %v", canvas.ID),
		"data":    canvas,
	})
}

func Websocket(c *gin.Context) {
	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	var paramId string = c.Param("id")
	var id uint64 = 0
	if paramId != "" {
		var err error
		id, err = strconv.ParseUint(paramId, 10, 64)
		if err != nil {
			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
			return
		}
	}

	conn := websocket_service.Connect(c)

	websocket_service.AddConnection(id, userUuid, conn)

	for {
		message := &websocket_service.CanvasMessage{}
		err := conn.ReadJSON(&message)
		if err != nil {
			logging.LogInfo("WebSocket", "Error reading message from websocket connection, closing connection", err)
		}

		response := &websocket_service.CanvasMessage{Event: message.Event, Email: message.Email, Data: message.Data}
		websocket_service.WriteToCanvasConnections(id, conn, response)
	}
}
