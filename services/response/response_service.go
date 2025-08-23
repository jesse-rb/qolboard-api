package response_service

import (
	"maps"
	model "qolboard-api/models"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jesse-rb/imissphp-go"
)

func BuildResponse(node any) any {
	t := reflect.TypeOf(node)
	v := reflect.ValueOf(node)

	if respondable, ok := node.(model.Respondable); ok {
		// var resp map[string]any
		resp := respondable.Response()

		// Iterate over node properties
		for i := 0; i < v.NumField(); i++ {
			// Get child node
			child := v.Field(i).Interface()
			jsonTag := t.Field(i).Tag.Get("json")

			_t := reflect.TypeOf(child)
			_v := reflect.ValueOf(child)

			// If child node is struct, build it's response
			if _t.Kind() == reflect.Struct {
				childResp := BuildResponse(child)

				if childResp != nil && jsonTag != "" {
					resp[jsonTag] = childResp
				}
			}

			// If child node is a slice, iterate over it to build its response
			if _t.Kind() == reflect.Slice {
				var childResp []any
				for j := 0; j < _v.Len(); j++ {
					item := _v.Index(j).Interface()
					itemResp := BuildResponse(item)
					if itemResp != nil {
						childResp = append(childResp, itemResp)
					}
				}

				if len(childResp) > 0 {
					resp[jsonTag] = childResp
				}
			}
		}

		return resp
	} else if t.Kind() == reflect.Slice {
		var resp []any
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			itemResp := BuildResponse(item)

			if itemResp != nil {
				resp = append(resp, itemResp)
			}
		}
		return resp
	} else {
		return nil
	}
}

func SetJSON(c *gin.Context, value any) {
	c.Set("response", imissphp.ToMap(value))
}

func MergeJSON(c *gin.Context, toMerge any) {
	response := GetJSON(c)

	maps.Copy(response, imissphp.ToMap(toMerge))

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
	code := GetCode(c)
	var response gin.H = GetJSON(c)

	c.JSON(code, response)
}

func Abort(c *gin.Context) {
	Response(c)
	panic(1)
}
