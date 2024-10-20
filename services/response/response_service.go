package response_service

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "response_service", log.Lshortfile+log.Ldate)
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "response_service", log.Lshortfile+log.Ldate)

// Thank you gpt
func toGinH(data any) gin.H {
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return gin.H{}
	}

	// Unmarshal the JSON into a map
	var result gin.H
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return gin.H{}
	}

	return result
}

func SetJSON(c *gin.Context, value any) {
	c.Set("response", toGinH(value))
}

func MergeJSON(c *gin.Context, toMerge any) {
	response := GetJSON(c)

	for k, v := range toGinH(toMerge) {
		response[k] = v
	}

	c.Set("response", response)
}

func SetCode(c *gin.Context, value int) {
	c.Set("code", value)
}

func GetJSON(c *gin.Context) gin.H {
	response, exists := c.Get("response")
	if !exists {
		response = gin.H{}
	}
	return response.(gin.H)
}

func GetCode(c *gin.Context) int {
	code, exists := c.Get("code")
	if !exists {
		code = 200
	}
	return code.(int)
}