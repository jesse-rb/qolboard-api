package error_service

import (
	"errors"
	"fmt"
	"qolboard-api/services/logging"
	trivial_service "qolboard-api/services/trivial"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type ErrorMeta struct {
	Field    string
	Value    string
	Resource string
	Code     int
}

type Error struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Value   string `json:"value"`
}

func SetUpValidator() {
	// Register go validator customizations
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Return `json` tag field name instead of internal go Struct field name
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			// skip if tag key says it should be ignored
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

func ValidationError(c *gin.Context, err error) {
	c.Error(err).SetType(gin.ErrorTypeBind)
}

func PublicError(c *gin.Context, message string, code int, field string, value string, resource string) {
	err := c.Error(errors.New(message))
	err.SetType(gin.ErrorTypePublic)
	err.SetMeta(ErrorMeta{
		Field:    field,
		Value:    value,
		Resource: resource,
		Code:     code,
	})
}

func InternalError(c *gin.Context, message string) {
	err := c.Error(errors.New(message))
	err.SetType(gin.ErrorTypePrivate)
	err.SetMeta(ErrorMeta{
		Field:    "",
		Value:    "",
		Resource: "",
		Code:     500,
	})
}

func HandleGinError(err gin.Error) (code int, formatted []*Error) {
	logging.LogInfo("error service", "We have encountered an error.", gin.H{
		"error": err.Error(),
	})

	if isValidationError(err) {
		code, formatted = newValidationErrors(err)
	} else if err.IsType(gin.ErrorTypePublic) {
		code, formatted = newPublicError(err)
	} else { // Default to internal server error (error type private)
		code, formatted = newInternalServerError()
	}

	return code, formatted
}

func isValidationError(err gin.Error) bool {
	_, ok := err.Err.(validator.ValidationErrors)
	return ok
}

func newPublicError(err gin.Error) (int, []*Error) {
	formatted := make([]*Error, 0)

	errorMeta, ok := err.Meta.(ErrorMeta)
	if !ok {
		return newInternalServerError()
	}

	var code int = errorMeta.Code

	var newError *Error = &Error{
		Message: err.Error(),
		Field:   errorMeta.Field,
		Value:   errorMeta.Value,
	}

	formatted = append(formatted, newError)

	return code, formatted
}

func newValidationErrors(err gin.Error) (int, []*Error) {
	formatted := make([]*Error, 0)

	validationErrors, ok := err.Err.(validator.ValidationErrors)
	if ok { // Check if validation error
		for _, v := range validationErrors {
			var field string = v.Field()
			var value string = fmt.Sprintf("%v", v.Value())
			var fieldEnglish string = trivial_service.UcFirst(strings.Join(strings.Split(field, "_"), " "))
			var tag string = v.Tag()

			// Default validation error msg
			var message string = fmt.Sprintf("field: %s with value: %s failed validation: %s", fieldEnglish, value, tag)

			// Detailed validation error msg
			if v.Tag() == "required" {
				message = fmt.Sprintf("%s is a required field.", fieldEnglish)
			}
			if v.Tag() == "email" {
				message = fmt.Sprintf("%s must be a valid email address.", fieldEnglish)
			}
			if v.Tag() == "oneof" {
				message = fmt.Sprintf("%s must of one of: %s", fieldEnglish, strings.Join(strings.Split(v.Param(), " "), ", "))
			}
			if v.Tag() == "lte" {
				message = fmt.Sprintf("%s must be less than or equal to %s", fieldEnglish, v.Param())
			}
			if v.Tag() == "gte" {
				message = fmt.Sprintf("%s must be greater than or equal to %s", fieldEnglish, v.Param())
			}

			var newError *Error = &Error{
				Message: message,
				Field:   field,
				Value:   value,
			}

			formatted = append(formatted, newError)
		}
	}

	return 422, formatted
}

func newInternalServerError() (int, []*Error) {
	formatted := make([]*Error, 0)

	var newError *Error = &Error{
		Message: "Sorry, something went wrong. :(",
		Field:   "",
		Value:   "",
	}

	formatted = append(formatted, newError)

	return 500, formatted
}
