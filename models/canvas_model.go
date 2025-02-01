package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	CanvasModeDraw   = "draw"
	CanvasModeGrab   = "grab"
	CanvasModePan    = "pan"
	CanvasModeRemove = "remove"
)

type Canvas struct {
	Model
	UserUuid               string                    `json:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;index"`
	CanvasData             datatypes.JSON            `json:"canvasData"`
	CanvasSharedAccess     []*CanvasSharedAccess     `json:"canvas_shared_accesses"`
	CanvasSharedInvitation []*CanvasSharedInvitation `json:"canvas_shared_invitations"`
	User                   *User                     `json:"user"`
}

type SerializedCanvas struct{} // TODO...

type CanvasData struct {
	// Width			*int	`json:"width" binding:"required"`
	// Height			*int	`json:"height" binding:"required"`
	Name            string         `json:"name" binding:"required"`
	ActiveMode      string         `json:"activeMode" binding:"required"`
	MouseDown       bool           `json:"mouseDown"`
	MouseX          *int           `json:"mouseX" binding:"required"`
	MouseY          *int           `json:"mouseY" binding:"required"`
	PrevMouseX      *int           `json:"prevMouseX" binding:"required"`
	PrevMouseY      *int           `json:"prevMouseY" binding:"required"`
	XPan            *int           `json:"xPan" binding:"required"`
	YPan            *int           `json:"yPan" binding:"required"`
	BackgroundColor string         `json:"backgroundColor" binding:"required"`
	PieceSettings   *PieceSettings `json:"pieceSettings" binding:"required"`
	Zoom            *float64       `json:"zoom" binding:"required"`
	ZoomDx          *float64       `json:"zoomDx"`
	ZoomDy          *float64       `json:"zoomDy"`
	RulerSettings   RulerSettings  `json:"rulerSettings"`
	PiecesManager   PiecesManager  `json:"piecesManager" binding:"required"`
}

type PieceSettings struct {
	Size   int    `json:"size" binding:"required"`
	Coloer string `json:"color" binding:"required"`
}

type RulerSettings struct {
	ShowUnits bool `json:"show_units"`
	ShowLines bool `json:"show_lines"`
}

type PiecesManager struct {
	Pieces     []*PieceData `json:"pieces"`
	LeftMost   *float64     `json:"leftMost" binding:"required"`
	RightMost  *float64     `json:"rightMost" binding:"required"`
	TopMost    *float64     `json:"topMost" binding:"required"`
	BottomMost *float64     `json:"bottomMost" binding:"required"`
}

type PieceData struct {
	Settings   *PieceSettings `json:"settings" binding:"required"`
	Path       string         `json:"path" binding:"required"`
	Move       DOMMatrixs     `json:"move" binding:"required"`
	Pan        DOMMatrixs     `json:"pan" binding:"required"`
	LeftMost   *float64       `json:"leftMost" binding:"required"`
	RightMost  *float64       `json:"rightMost" binding:"required"`
	TopMost    *float64       `json:"topMost" binding:"required"`
	BottomMost *float64       `json:"bottomMost" binding:"required"`
}

type DOMMatrixs struct {
	A   float64 `json:"a" binding:"required"`
	B   float64 `json:"b" binding:"required"`
	C   float64 `json:"c" binding:"required"`
	D   float64 `json:"d" binding:"required"`
	E   float64 `json:"e" binding:"required"`
	F   float64 `json:"f" binding:"required"`
	M11 float64 `json:"m11" binding:"required"`
	M12 float64 `json:"m12" binding:"required"`
	M13 float64 `json:"m13" binding:"required"`
	M14 float64 `json:"m14" binding:"required"`
	M21 float64 `json:"m21" binding:"required"`
	M22 float64 `json:"m22" binding:"required"`
	M23 float64 `json:"m23" binding:"required"`
	M24 float64 `json:"m24" binding:"required"`
	M31 float64 `json:"m31" binding:"required"`
	M32 float64 `json:"m32" binding:"required"`
	M33 float64 `json:"m33" binding:"required"`
	M34 float64 `json:"m34" binding:"required"`
	M41 float64 `json:"m41" binding:"required"`
	M42 float64 `json:"m42" binding:"required"`
	M43 float64 `json:"m43" binding:"required"`
	M44 float64 `json:"m44" binding:"required"`
}

func CanvasBelongsToUser(userUuid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas.user_uuid", userUuid)
	}
}
