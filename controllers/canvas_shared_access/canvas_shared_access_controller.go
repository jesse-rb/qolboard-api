package canvas_shared_access_controller

import (
	"fmt"
	"net/http"
	database_config "qolboard-api/config/database"
	controller "qolboard-api/controllers"
	model "qolboard-api/models"
	canvas_shared_access_model "qolboard-api/models/canvas_shared_access"
	error_service "qolboard-api/services/error"
	generator_service "qolboard-api/services/generator"
	relations_service "qolboard-api/services/relations"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IndexParams struct {
	controller.IndexParams
}

func Index(c *gin.Context) {
	// Get query params
	params := IndexParams{
		IndexParams: controller.IndexParams{
			Page:  1,
			Limit: 100,
			With:  make([]string, 0),
		},
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	tx, err := database_config.DB(c)
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	csa, error := canvas_shared_access_model.GetAll(tx, params.Limit, params.Page)
	if error != nil {
		error_service.InternalError(c, error.Error())
		return
	}

	err = relations_service.LoadBatch(tx, model.CanvasSharedAccessRelations, csa, params.With)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"data": generator_service.BuildResponse(csa),
	})

	resp := generator_service.BuildResponse(csa)

	response_service.SetJSON(c, gin.H{
		"data": resp,
	})

	tx.Commit()
}

func Delete(c *gin.Context) {
	// Parse id
	paramId := c.Param("canvas_shared_access_id")
	id, err := strconv.ParseUint(paramId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas_shared_access")
		return
	}

	tx, err := database_config.DB(c)
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	csa := model.CanvasSharedAccess{}
	csa.ID = id

	err = csa.Delete(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	resp := generator_service.BuildResponse(csa)

	response_service.SetJSON(c, gin.H{
		"message": fmt.Sprintf("Successfully deleted canvas shared access with id %v", csa.ID),
		"data":    resp,
	})

	tx.Commit()
}
