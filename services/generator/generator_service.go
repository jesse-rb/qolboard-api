package generator_service

import (
	model "qolboard-api/models"
	"reflect"
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
