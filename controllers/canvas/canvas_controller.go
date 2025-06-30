package canvas_controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	database_config "qolboard-api/config/database"
	controller "qolboard-api/controllers"
	model "qolboard-api/models"
	canvas_model "qolboard-api/models/canvas"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	generator_service "qolboard-api/services/generator"
	relations_service "qolboard-api/services/relations"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type getParams struct {
	controller.GetParams
}

type indexParams struct {
	controller.IndexParams
}

func Index(c *gin.Context) {
	var params indexParams = indexParams{
		IndexParams: controller.IndexParams{
			Page:  1,
			Limit: 100,
			With:  make([]string, 0),
		},
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		error_service.ValidationError(c, err)
		return
	}

	tx, err := database_config.DB(c)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	defer tx.Commit()

	canvases, err := canvas_model.GetAll(tx, params.Limit, params.Page)
	if err != nil {
		tx.Rollback()
		error_service.InternalError(c, err.Error())
		return
	}

	err = relations_service.LoadBatch(tx, model.CanvasRelations, canvases, params.With)
	if err != nil {
		tx.Rollback()
		error_service.InternalError(c, err.Error())
		return
	}

	resp := generator_service.BuildResponse(canvases)

	response_service.SetJSON(c, gin.H{
		"data": resp,
	})
}

func Get(c *gin.Context) {
	var params getParams = getParams{
		GetParams: controller.GetParams{
			With: make([]string, 0),
		},
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		error_service.ValidationError(c, err)
		return
	}

	// claims := auth_service.GetClaims(c)

	var paramId string = c.Param("canvas_id")
	id, err := strconv.ParseUint(paramId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Canvas id must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas")
		return
	}

	tx, err := database_config.DB(c)
	defer tx.Commit()
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	canvas, err := canvas_model.Get(tx, id)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	err = relations_service.Load(tx, model.CanvasRelations, canvas, params.With)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	resp := generator_service.BuildResponse(*canvas)

	response_service.SetJSON(c, map[string]any{
		"data": resp,
	})
}

func Save(c *gin.Context) {
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

	var canvasData canvas_model.CanvasData
	if err := c.ShouldBindJSON(&canvasData); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	canvasDataJson, err := json.Marshal(canvasData)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	tx, err := database_config.DB(c)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}
	defer tx.Commit()

	canvas := &model.Canvas{}
	canvas.ID = id
	canvas.CanvasData = canvasDataJson
	canvas.UserUuid = userUuid

	err = canvas.Save(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	// var result *gorm.DB
	// if id > 0 {
	// 	// Update
	// 	canvas.ID = id
	//
	// 	query := db.Connection
	// 	model.Canvas{}.LeftJoinCanvasSharedAccessOnUser(query, userUuid)
	// 	model.Canvas{}.BelongsToUser(query, userUuid)
	// 	query.Or(model.CanvasSharedAccess{}.BelongsToCanvas(query, nil))
	// 	result = query.First(&canvas)
	//
	// 	if result.Error != nil {
	// 		if result.Error == gorm.ErrRecordNotFound {
	// 			error_service.PublicError(c, "Canvas not found", http.StatusNotFound, "id", paramId, "canvas")
	// 			return
	// 		}
	//
	// 		error_service.InternalError(c, result.Error.Error())
	// 		return
	// 	}
	//
	// 	canvas.CanvasData = canvasDataJson
	// 	result = db.Connection.Save(&canvas)
	// } else {
	// 	canvas.UserUuid = userUuid
	// 	canvas.CanvasData = canvasDataJson
	//
	// 	query := db.Connection
	// 	model.Canvas{}.BelongsToUser(query, userUuid)
	// 	result = query.Save(&canvas)
	// }
	//
	// if result.Error != nil {
	// 	error_service.InternalError(c, result.Error.Error())
	// 	return
	// }
	//
	// db.Connection.
	// 	Preload("User").
	// 	Preload("CanvasSharedAccess").
	// 	Preload("CanvasSharedAccess.User").
	// 	Preload("CanvasSharedInvitation").
	// 	First(&canvas)

	response_service.SetJSON(c, gin.H{
		"msg":    fmt.Sprintf("Successfully saved canvas with id: %v", canvas.ID),
		"canvas": canvas,
	})
}

//
//	func Delete(c *gin.Context) {
//		db := database_config.GetDatabase()
//
//		claims := auth_service.GetClaims(c)
//		userUuid := claims.Subject
//
//		var paramId string = c.Param("canvas_id")
//		var id uint64 = 0
//		var err error = nil
//		if paramId != "" {
//			id, err = strconv.ParseUint(paramId, 10, 64)
//			if err != nil {
//				error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
//				return
//			}
//		}
//
//		var canvas model.Canvas
//
//		canvas.ID = id
//
//		query := db.Connection
//		model.Canvas{}.BelongsToUser(query, userUuid)
//		query.First(&canvas, id)
//
//		query = db.Connection
//		model.Canvas{}.BelongsToUser(query, userUuid)
//		query.Select("CanvasSharedAccess", "CanvasSharedInvitation")
//		result := query.Delete(&canvas, id)
//
//		if result.Error != nil {
//			error_service.InternalError(c, result.Error.Error())
//			return
//		}
//
//		response_service.SetJSON(c, gin.H{
//			"message": fmt.Sprintf("Successfully deleted canvas shared invitation with id %v", canvas.ID),
//			"data":    canvas,
//		})
//	}
// func Websocket(c *gin.Context) {
// 	claims := auth_service.GetClaims(c)
// 	userUuid := claims.Subject
//
// 	// Validate canvas id param
// 	var paramId string = c.Param("id")
// 	var id uint64 = 0
// 	if paramId != "" {
// 		var err error
// 		id, err = strconv.ParseUint(paramId, 10, 64)
// 		if err != nil {
// 			error_service.PublicError(c, "Canvas id must be an integer", http.StatusUnprocessableEntity, "canvas_id", paramId, "canvas")
// 			return
// 		}
// 	}
//
// 	db := database_config.GetDatabase()
//
// 	// Validate user owns canvas or has access to canvas
// 	var canvas model.Canvas
// 	canvas.ID = id
//
// 	query := db.Connection
// 	model.Canvas{}.LeftJoinCanvasSharedAccessOnUser(query, userUuid)
// 	model.Canvas{}.BelongsToUser(query, userUuid)
// 	query.Or(model.CanvasSharedAccess{}.BelongsToCanvas(query, &id))
// 	result := query.First(&canvas)
//
// 	if result.Error != nil {
// 		if result.Error == gorm.ErrRecordNotFound {
// 			error_service.PublicError(c, "Could not find canvas", http.StatusNotFound, "id", paramId, "canvas")
// 		} else {
// 			error_service.InternalError(c, result.Error.Error())
// 		}
// 		return
// 	}
//
// 	conn := websocket_service.Connect(c)
//
// 	websocket_service.AddConnection(id, userUuid, conn)
//
// 	for {
// 		message := &websocket_service.CanvasMessage{}
// 		err := conn.ReadJSON(&message)
// 		if err != nil {
// 			logging.LogInfo("WebSocket", "Error reading message from websocket connection, closing connection", err)
// 		}
//
// 		response := &websocket_service.CanvasMessage{Event: message.Event, Email: message.Email, Data: message.Data}
// 		websocket_service.WriteToCanvasConnections(id, conn, response)
// 	}
// }
