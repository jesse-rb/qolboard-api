package canvas_shared_invitation_controller

import (
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IndexQuery struct {
	CanvasId uint64 `form:"canvas_id"`
	Page     uint64 `form:"page"`
	Limit    uint64 `form:"limit"`
}

func Create(c *gin.Context) {
	db := database_config.GetDatabase()

	var claims auth_service.Claims = *auth_service.GetClaims(c)

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

	canvasSharedInvitation, err = model.NewCanvasSharedInvitation(claims.Subject, canvasId)
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

	response_service.SetJSON(c, gin.H{
		"data": canvasSharedInvitation.Response(),
	})
}

func Index(c *gin.Context) {
	// Get user claims
	claims := auth_service.GetClaims(c)

	// Get query params
	var queryValues IndexQuery
	if err := c.ShouldBindQuery(&queryValues); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	// Query with filters
	db := database_config.GetDatabase()

	query := db.Connection.Model(&model.CanvasSharedInvitation{})

	// User UUID
	query.Scopes(model.CanvasSharedInvitationBelongsToUser(claims.Subject))

	// Canvas ID
	if queryValues.CanvasId > 0 {
		query.Scopes(model.CanvasSharedInvitationBelongsToCanvas(queryValues.CanvasId))
	}

	// Pagination
	page := 0
	limit := 100
	if queryValues.Page > 0 {
		page = int(queryValues.Page)
	}
	if queryValues.Limit > 0 {
		limit = min(limit, int(queryValues.Limit))
	}

	query.Limit(limit)
	query.Offset(limit * page)

	var data []*model.CanvasSharedInvitation
	query.Find(&data)

	response_service.SetJSON(c, gin.H{
		"data": data,
	})
}

func AcceptInvite(c *gin.Context) {
}
