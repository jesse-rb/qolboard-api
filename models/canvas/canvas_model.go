package canvas_model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Canvas struct {
	gorm.Model
	UserEmail string
	CanvasData datatypes.JSON
}