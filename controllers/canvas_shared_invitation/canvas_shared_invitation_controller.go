package canvas_shared_invitation_controller

import (
	"net/http"
	error_service "qolboard-api/services/error"

	"github.com/gin-gonic/gin"
)

func Create(c *gin.Context) {
	// db := database_config.GetDatabase()

	var paramCanvasId string = c.Param("canvas_id")
	// var canvasId uint64 = 0
	var err error = nil

	// Parse params
	// canvasId, err = strconv.ParseUint(paramCanvasId, 10, 64)
	if err != nil {
		error_service.PublicError(c, "Canvas id must be a valid integer", http.StatusUnprocessableEntity, "canvas_id", paramCanvasId, "canvas")
		return
	}
}
