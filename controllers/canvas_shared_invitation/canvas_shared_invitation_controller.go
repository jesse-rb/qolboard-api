package canvas_shared_invitation_controller

import (
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	canvas_shared_invitation_model "qolboard-api/models/canvas_shared_invitation"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IndexQuery struct {
	CanvasId uint64   `form:"canvas_id"`
	Page     uint64   `form:"page"`
	Limit    uint64   `form:"limit"`
	With     []string `form:"with[]" binding:"dive,oneof=canvas canvas_shared_access"`
}

func Create(c *gin.Context) {
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

	canvasSharedInvitation, err = canvas_shared_invitation_model.NewCanvasSharedInvitation(claims.Subject, canvasId)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	tx, err := database_config.DB(c)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	err = canvasSharedInvitation.Save(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	tx.Commit()

	canvasSharedInvitation.Response()

	response_service.SetJSON(c, gin.H{
		"data": canvasSharedInvitation,
	})
}

func Index(c *gin.Context) {
	// Get user claims
	// claims := auth_service.GetClaims(c)

	// Get query params
	var queryValues IndexQuery
	if err := c.ShouldBindQuery(&queryValues); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	// Query with filters
	tx, err := database_config.DB(c)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	data, err := canvas_shared_invitation_model.GetAllForCanvas(tx, queryValues.CanvasId)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	tx.Commit()

	// Format response
	for i := range data {
		data[i].Response()
	}

	response_service.SetJSON(c, gin.H{
		"data": data,
	})
}

// func Delete(c *gin.Context) {
// 	claims := auth_service.GetClaims(c)
//
// 	// Parse id
// 	paramId := c.Param("canvas_shared_invitation_id")
// 	id, err := strconv.ParseUint(paramId, 10, 64)
// 	if err != nil {
// 		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas_shared_invitation")
// 		return
// 	}
//
// 	db := database_config.GetDatabase()
//
// 	// Find record
// 	var sharedInvitation model.CanvasSharedInvitation
// 	db.Connection.
// 		Scopes(model.CanvasSharedInvitationBelongsToUser(claims.Subject)).
// 		First(&sharedInvitation, id)
//
// 	// Delete record
// 	result := db.Connection.
// 		Scopes(model.CanvasSharedInvitationBelongsToUser(claims.Subject)).
// 		Delete(&sharedInvitation, id)
// 	if result.Error != nil {
// 		error_service.InternalError(c, result.Error.Error())
// 		return
// 	}
//
// 	response_service.SetJSON(c, gin.H{
// 		"message": fmt.Sprintf("Successfully deleted shared invitiation with id %v", sharedInvitation.ID),
// 		"data":    sharedInvitation,
// 	})
// }
//
// func AcceptInvite(c *gin.Context) {
// 	claims := auth_service.GetClaims(c)
//
// 	paramCanvasId := c.Param("canvas_id")
// 	paramCode := c.Param("code")
//
// 	// Parse & validate params
// 	canvasId, err := strconv.ParseUint(paramCanvasId, 10, 64)
// 	if err != nil {
// 		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramCanvasId, "canvas")
// 		return
// 	}
//
// 	// Find shared invitation by code and canvas id
// 	db := database_config.GetDatabase()
//
// 	var sharedInvitation model.CanvasSharedInvitation
// 	result := db.Connection.Scopes(model.CanvasSharedInvitationBelongsToCanvas(canvasId)).
// 		Where(&model.CanvasSharedInvitation{Code: paramCode}).
// 		First(&sharedInvitation)
//
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			link := c.Request.URL.String()
// 			error_service.PublicError(c, "Could not find this invite link", http.StatusNotFound, "link", link, "canvas_shared_invitation")
// 		} else {
// 			error_service.InternalError(c, result.Error.Error())
// 		}
// 		return
// 	}
//
// 	// Check to ensure we do not create a "shared access" for the canvas owner
// 	if sharedInvitation.UserUuid != claims.Subject {
// 		// Create shared access
// 		var sharedAccess model.CanvasSharedAccess = model.CanvasSharedAccess{
// 			UserUuid:                 claims.Subject,
// 			CanvasId:                 canvasId,
// 			CanvasSharedInvitationId: sharedInvitation.ID,
// 		}
//
// 		db.Connection.Save(&sharedAccess)
// 	}
//
// 	// Redirect to canvas
// 	appHost := os.Getenv("APP_HOST")
// 	locatoin := fmt.Sprintf("%s/canvas/%v", appHost, canvasId)
// 	c.Redirect(http.StatusFound, locatoin)
// }
