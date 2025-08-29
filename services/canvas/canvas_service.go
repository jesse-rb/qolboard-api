package canvas_service

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const (
	CanvasModeDraw   = "draw"
	CanvasModeGrab   = "grab"
	CanvasModePan    = "pan"
	CanvasModeRemove = "remove"
)

type CanvasData struct {
	Name            string         `json:"name" binding:"required"`
	BackgroundColor string         `json:"backgroundColor" binding:"required"`
	PieceSettings   *PieceSettings `json:"pieceSettings" binding:"required"`
	RulerSettings   RulerSettings  `json:"rulerSettings"`
	PiecesManager   *PiecesManager `json:"piecesManager" binding:"required"`
}

type PieceSettings struct {
	Size   int    `json:"size" binding:"required"`
	Coloer string `json:"color" binding:"required"`
}

type RulerSettings struct {
	ShowUnits bool `json:"showUnits"`
	ShowLines bool `json:"showLines"`
}

type PiecesManager struct {
	Pieces     []*PieceData `json:"pieces"`
	LeftMost   *float64     `json:"leftMost" binding:"required"`
	RightMost  *float64     `json:"rightMost" binding:"required"`
	TopMost    *float64     `json:"topMost" binding:"required"`
	BottomMost *float64     `json:"bottomMost" binding:"required"`
}

type PieceData struct {
	Settings *PieceSettings `json:"settings" binding:"required"`
	Path     string         `json:"path" binding:"required"`
	Move     DOMMatrixs     `json:"move" binding:"required"`
	// Pan        DOMMatrixs     `json:"pan" binding:"required"`
	LeftMost   *float64 `json:"leftMost" binding:"required"`
	RightMost  *float64 `json:"rightMost" binding:"required"`
	TopMost    *float64 `json:"topMost" binding:"required"`
	BottomMost *float64 `json:"bottomMost" binding:"required"`
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

func (canvasData *CanvasData) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan CanvasData: %v", value)
	}

	return json.Unmarshal(bytes, canvasData)
}

func (c CanvasData) Value() (driver.Value, error) {
	return json.Marshal(c)
}
