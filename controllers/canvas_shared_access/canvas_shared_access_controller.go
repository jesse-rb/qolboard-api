package canvas_shared_access_controller

import (
	"errors"
	"fmt"
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IndexQuery struct {
	CanvasId uint64   `form:"canvas_id"`
	Page     uint64   `form:"page"`
	Limit    uint64   `form:"limit"`
	With     []string `from:"with"`
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

	var data []*model.CanvasSharedAccess

	// Query with filters
	db := database_config.GetDatabase()

	// User UUID
	query := db.Connection.Scopes(model.CanvasSharedAccessBelongsToUserThroughCanvas(claims.Subject))

	// Canvas ID
	if queryValues.CanvasId > 0 {
		query.Scopes(model.CanvasSharedAccessBelongsToCanvas(queryValues.CanvasId))
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

	// With
	if slices.Contains(queryValues.With, "user") {
		query.Preload("User")
		query.Preload("Canvas")
	}
	query.Preload("User")

	query.Limit(limit)
	query.Offset(limit * page)

	query.Find(&data)

	response_service.SetJSON(c, gin.H{
		"data": data,
	})
}

func Delete(c *gin.Context) {
	claims := auth_service.GetClaims(c)

	// Parse id
	paramId := c.Param("canvas_shared_access_id")
	id, err := strconv.ParseUint(paramId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas_shared_access")
		return
	}

	db := database_config.GetDatabase()

	// Find record
	var sharedAccess model.CanvasSharedAccess
	result := db.Connection.
		Scopes(model.CanvasSharedAccessBelongsToUser(claims.Subject)).
		First(&sharedAccess, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			error_service.PublicError(c, "Could not find canvas shared access", http.StatusNotFound, "id", paramId, "canvas_shared_access")
		} else {
			error_service.InternalError(c, result.Error.Error())
		}
		return
	}

	// Delete record
	result = db.Connection.
		Scopes(model.CanvasSharedAccessBelongsToUser(claims.Subject)).
		Delete(&sharedAccess, id)
	if result.Error != nil {
		error_service.InternalError(c, result.Error.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": fmt.Sprintf("Successfully deleted canvas shared access with id %v", sharedAccess.ID),
		"data":    sharedAccess.CanvasSharedInvitation.CanvasSharedAccess,
	})
}
