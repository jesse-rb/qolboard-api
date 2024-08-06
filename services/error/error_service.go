package error_service

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "error_service", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "error_service", log.Lshortfile+log.Ldate);

type ErrorMeta struct {
	Field string
	Value string
	Resource string
	Code int
}

type Error struct {
	Message string `json:"message"`
	Field string `json:"field"`
	Value string `json:"value"`
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

func PublicError(c *gin.Context, message string, code int, field string, value string, resource string) {
	err := c.Error(fmt.Errorf(message))
	err.SetType(gin.ErrorTypePublic)
	err.SetMeta(ErrorMeta{
		Field: field,
		Value: value,
		Resource: resource,
		Code: code,
	})
}

func InternalError(c *gin.Context, message string) {
	err := c.Error(fmt.Errorf(message))
	err.SetType(gin.ErrorTypePrivate)
	err.SetMeta(ErrorMeta{
		Field: "",
		Value: "",
		Resource: "",
		Code: 500,
	})
}

func HandleGinError(err gin.Error) (code int, formatted []*Error) {

	infoLogger.Log("HandleGinError", "Handling gin error", gin.H{
		"error": err.Error(),
		"error_type_any": err.IsType(gin.ErrorTypeAny),
		"error_type_public": err.IsType(gin.ErrorTypePublic),
		"error_type_private": err.IsType(gin.ErrorTypePrivate),
		"error_type_validation": isValidationError(err),
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

func isValidationError(err gin.Error) (bool) {
	_, ok := err.Err.(validator.ValidationErrors)
	return ok
}

func newPublicError(err gin.Error) (int, []*Error) {
	var formatted = make([]*Error, 0)

	errorMeta, ok := err.Meta.(ErrorMeta);
	if !ok {
		return newInternalServerError()
	}

	var code int = errorMeta.Code

	var newError *Error = &Error{
		Message: err.Error(),
		Field: errorMeta.Field,
		Value: errorMeta.Value,
	}

	formatted = append(formatted, newError)

	return code, formatted
}

func newValidationErrors(err gin.Error) (int, []*Error) {
	var formatted = make([]*Error, 0)

	validationErrors, ok := err.Err.(validator.ValidationErrors)
	if ok { // Check if validation error
		for _, v := range validationErrors {
			var field string = v.Field()
			var value string = fmt.Sprintf("%s", v.Value())
			var tag string = v.Tag()
			var message string = fmt.Sprintf("field: %s with value: %s failed validation: %s", field, value, tag);
			var newError *Error = &Error{
				Message: message,
				Field: field,
				Value: value,
			}
			
			formatted = append(formatted, newError)
		}
	}

	return 422, formatted
}

func newInternalServerError() (int, []*Error) {
	var formatted = make([]*Error, 0)

	var newError *Error = &Error{
		Message: "Sorry, something went wrong. :(",
		Field: "",
		Value: "",
	}

	formatted = append(formatted, newError)

	return 500, formatted
}