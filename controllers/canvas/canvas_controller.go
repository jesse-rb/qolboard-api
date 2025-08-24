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
	relations_service "qolboard-api/services/relations"
	response_service "qolboard-api/services/response"
	websocket_service "qolboard-api/services/websocket"
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

	resp := response_service.BuildResponse(canvases)

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
		error_service.PublicError(c, "Could not find canvas", 404, "id", paramId, "canvas")
		tx.Rollback()
		return
	}

	err = relations_service.Load(tx, model.CanvasRelations, canvas, params.With)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	resp := response_service.BuildResponse(*canvas)

	response_service.SetJSON(c, map[string]any{
		"data": resp,
	})
}

func Save(c *gin.Context) {
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
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	canvas := &model.Canvas{}
	canvas.ID = id
	canvas.CanvasData = canvasDataJson

	err = canvas.Save(tx)
	if err != nil {
		error_service.PublicError(c, "Canvas not found", http.StatusNotFound, "canvas_id", paramId, "canvas")
		return
	}
	canvas, err = canvas_model.Get(tx, id)
	if err != nil {
		error_service.InternalError(c, err.Error())
	}

	err = relations_service.Load(tx, canvas.GetRelations(), canvas, []string{"user", "canvas_shared_invitations", "canvas_shared_accesses"})
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"msg":    fmt.Sprintf("Successfully saved canvas with id: %v", canvas.ID),
		"canvas": canvas,
	})
	tx.Commit()
}

func Delete(c *gin.Context) {
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

	canvas := model.Canvas{}
	canvas.ID = id
	canvas.UserUuid = userUuid

	tx, err := database_config.DB(c)
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	err = canvas.Delete(tx)
	if err != nil {
		error_service.PublicError(c, "Could not delete canvas", http.StatusNotFound, "canvas_id", paramId, "canvas")
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": fmt.Sprintf("Successfully deleted canvas with id %v", canvas.ID),
		"data":    response_service.BuildResponse(canvas),
	})

	tx.Commit()
}

func Websocket(c *gin.Context) {
	claims := auth_service.GetClaims(c)
	userUuid := claims.Subject

	// Parse query params
	var params getParams = getParams{
		GetParams: controller.GetParams{
			With: make([]string, 0),
		},
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		error_service.ValidationError(c, err)
		return
	}

	// Validate canvas id param
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

	tx, err := database_config.DB(c)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// Validate user owns canvas or has access to canvas
	canvas, err := canvas_model.Get(tx, id)
	if err != nil {
		error_service.PublicError(c, "Could not find canvas", http.StatusNotFound, "id", paramId, "canvas")
		return
	}

	chResume := make(chan *websocket_service.Client, 1)

	conn := websocket_service.Connect(c)
	websocket_service.Join(userUuid, canvas, conn, chResume)

	client := <-chResume

	// Go rotine for reading websocket messages
	go client.Reader(c)

	// Go rotine for writing websocket messages
	client.Writer()
}
