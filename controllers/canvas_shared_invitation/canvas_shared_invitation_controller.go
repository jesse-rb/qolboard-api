package canvas_shared_invitation_controller

import (
	"fmt"
	"net/http"
	"os"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	canvas_shared_invitation_model "qolboard-api/models/canvas_shared_invitation"
	auth_service "qolboard-api/services/auth"
	error_service "qolboard-api/services/error"
	generator_service "qolboard-api/services/generator"
	"qolboard-api/services/logging"
	relations_service "qolboard-api/services/relations"
	response_service "qolboard-api/services/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IndexQuery struct {
	CanvasId uint64   `form:"canvas_id"`
	Page     uint64   `form:"page"`
	Limit    uint64   `form:"limit"`
	With     []string `form:"with[]"`
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
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	err = canvasSharedInvitation.Save(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	tx.Commit()

	resp := generator_service.BuildResponse(*canvasSharedInvitation)

	response_service.SetJSON(c, gin.H{
		"data": resp,
	})
}

func Index(c *gin.Context) {
	// Get user claims
	// claims := auth_service.GetClaims(c)

	// Get query params
	var params IndexQuery
	if err := c.ShouldBindQuery(&params); err != nil {
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

	data, err := canvas_shared_invitation_model.GetAllForCanvas(tx, params.CanvasId)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	err = relations_service.LoadBatch(tx, model.CanvasSharedInvitationRelations, data, params.With)
	if err != nil {
		error_service.InternalError(c, err.Error())
		tx.Rollback()
		return
	}

	tx.Commit()

	resp := generator_service.BuildResponse(data)

	response_service.SetJSON(c, gin.H{
		"data": resp,
	})
}

func Delete(c *gin.Context) {
	// Parse id
	paramId := c.Param("canvas_shared_invitation_id")
	id, err := strconv.ParseUint(paramId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramId, "canvas_shared_invitation")
		return
	}

	tx, err := database_config.DB(c)
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	csi := model.CanvasSharedInvitation{}
	csi.ID = id

	debug := model.CanvasSharedInvitation{}
	debug.ID = id
	tx.Get(&debug, "SELECT * FROM canvas_shared_invitations WHERE id = $1 AND user_uuid = get_user_uuid() AND deleted_at IS NULL", id)
	logging.LogDebug("canvas_shared_invitation_controller", "Finding the csi", debug)

	err = csi.Delete(tx)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	response_service.SetJSON(c, gin.H{
		"message": "successfully deleted shared canvas link",
		"data":    generator_service.BuildResponse(csi),
	})

	tx.Commit()
}

func AcceptInvite(c *gin.Context) {
	claims := auth_service.GetClaims(c)

	paramCanvasId := c.Param("canvas_id")
	paramCode := c.Param("code")

	// Parse & validate params
	canvasId, err := strconv.ParseUint(paramCanvasId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Must be a valid integer", http.StatusUnprocessableEntity, "id", paramCanvasId, "canvas")
		return
	}

	// Find shared invitation by code and canvas id
	tx, err := database_config.DB(c)
	defer tx.Rollback()
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	csi, err := canvas_shared_invitation_model.GetByCode(tx, canvasId, paramCode)
	if err != nil {
		error_service.InternalError(c, err.Error())
		return
	}

	// Check to ensure we do not create a "shared access" for the canvas owner
	if csi.UserUuid != claims.Subject {
		// Create shared access
		var csa model.CanvasSharedAccess = model.CanvasSharedAccess{
			UserUuid:                 claims.Subject,
			CanvasId:                 canvasId,
			CanvasSharedInvitationId: csi.ID,
		}

		csa.Insert(tx)
	}

	tx.Commit()

	// Redirect to canvas
	appHost := os.Getenv("APP_HOST")
	locatoin := fmt.Sprintf("%s/canvas/%v", appHost, canvasId)
	c.Redirect(http.StatusFound, locatoin)
}
