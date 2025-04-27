package response_service

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/jesse-rb/imissphp-go"
)

// Thank you gpt
func _ToMap(data any) map[string]any {
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]any{}
	}

	// Unmarshal the JSON into a map
	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return map[string]any{}
	}

	return result
}

func SetJSON(c *gin.Context, value any) {
	c.Set("response", imissphp.ToMap(value))
}

func MergeJSON(c *gin.Context, toMerge any) {
	response := GetJSON(c)

	for k, v := range imissphp.ToMap(toMerge) {
		response[k] = v
	}

	c.Set("response", response)
}

func SetCode(c *gin.Context, value int) {
	c.Set("code", value)
}

func GetJSON(c *gin.Context) map[string]any {
	response, exists := c.Get("response")
	if !exists {
		response = map[string]any{}
	}
	return response.(map[string]any)
}

func GetCode(c *gin.Context) int {
	code, exists := c.Get("code")
	if !exists {
		code = 200
	}
	return code.(int)
}

func Response(c *gin.Context) {
	var code int = GetCode(c)
	var response gin.H = GetJSON(c)

	c.JSON(code, response)
}

func Abort(c *gin.Context) {
	Response(c)
	panic(1)
}
