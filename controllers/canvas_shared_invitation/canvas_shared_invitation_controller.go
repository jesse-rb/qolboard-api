package canvas_shared_invitation_controller

import (
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Create(c *gin.Context) {
	db := database_config.GetDatabase()

	var paramCanvasId string = c.Param("canvas_id")
	var canvasId uint64 = 0
	var err error = nil

	// Parse params
	canvasId, err = strconv.ParseUint(paramCanvasId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Canvas id must be a valid integer", http.StatusUnprocessableEntity, "canvas_id", paramCanvasId, "canvas")
		return
	}

	var canvasSharedInvitation *model.CanvasSharedInvitation

	canvasSharedInvitation, err = model.NewCanvasSharedInvitation(canvasId)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	result := db.Connection.
		Scopes(model.CanvasSharedInvitationBelongsToCanvas(canvasId)).
		Save(canvasSharedInvitation)

	if result.Error != nil {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetCode(c, 200)
	response_service.SetJSON(c, gin.H{
		"data": canvasSharedInvitation.Response(),
	})
}

type indexQuery struct {
	CanvasId uint64 `json:"canvas_id"`
	Page     uint64 `json:"page"`
	Limit    uint64 `json:"limit"`
}

func index(c *gin.Context) {
	var query *indexQuery

	if err := c.ShouldBindQuery(query); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	db := database_config.GetDatabase()

	// db.Connection.Scopes(funcs ...func(*gorm.DB) *gorm.DB)
}

func AcceptInvite(c *gin.Context) {
}
