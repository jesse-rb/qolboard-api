package error_middleware

import (
	"net/http"
	error_service "qolboard-api/services/error"
	response_service "qolboard-api/services/response"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Value   string `json:"value"`
}

func Run(c *gin.Context) {
	c.Next()

	var errors []*error_service.Error = make([]*error_service.Error, 0)
	var code int = 500

	internalServerErrors := c.Errors.ByType(gin.ErrorTypePrivate)
	validationErrors := c.Errors.ByType(gin.ErrorTypeBind)
	publicErrors := c.Errors.ByType(gin.ErrorTypePublic)

	for _, err := range internalServerErrors {
		var formatted []*error_service.Error
		_, formatted = error_service.HandleGinError(*err)

		errors = append(errors, formatted...)
	}
	for _, err := range validationErrors {
		var formatted []*error_service.Error
		_, formatted = error_service.HandleGinError(*err)

		errors = append(errors, formatted...)
	}
	for _, err := range publicErrors {
		var formatted []*error_service.Error
		code, formatted = error_service.HandleGinError(*err)

		errors = append(errors, formatted...)
	}

	if len(internalServerErrors) > 0 {
		response_service.SetCode(c, http.StatusInternalServerError)
	} else if len(validationErrors) > 0 {
		response_service.SetCode(c, http.StatusUnprocessableEntity)
	} else if len(publicErrors) > 0 {
		response_service.SetCode(c, code)
	}

	response_service.MergeJSON(c, gin.H{
		"errors": errors,
	})
}

