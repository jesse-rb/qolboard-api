package canvas_model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
    CanvasModeDraw		= "draw"
    CanvasModeGrab      = "grab"
    CanvasModePan		= "pan"
	CanvasModeRemove	= "remove"
)

type Canvas struct {
	gorm.Model
	UserEmail string
	CanvasData datatypes.JSON
}

type SerializedCanvas struct {} // TODO...

type CanvasData struct {	
	Width			*int	`json:"width" binding:"required"`
	Height			*int	`json:"height" binding:"required"`
	ActiveMode		string	`json:"activeMode" binding:"required"`
	MouseDown		bool	`json:"mouseDown"`
	MouseX			*int	`json:"mouseX" binding:"required"`
	MouseY			*int	`json:"mouseY" binding:"required"`
	PrevMouseX		*int	`json:"prevMouseX" binding:"required"`
	PrevMouseY		*int	`json:"prevMouseY" binding:"required"`
	XPan			*int	`json:"xPan" binding:"required"`
	YPan			*int	`json:"yPan" binding:"required"`
	BackgroundColor	string	`json:"backgroundColor" binding:"required"`
	PieceSettings 	*PieceSettings `json:"pieceSettings" binding:"required"`
	Zoom			*int	`json:"zoom" binding:"required"`
	ZoomDx			*int	`json:"zoomDx"`
	ZoomDy			*int	`json:"zoomDy"`
	PiecesManager	PiecesManager `json:"piecesManager" binding:"required"`
}

type PieceSettings struct {
	Size	int `json:"size" binding:"required"`
	Coloer	string `json:"color" binding:"required"`
}

type PiecesManager struct {
	Pieces []*PieceData `json:"pieces"`
}

type PieceData struct {
	Settings	*PieceSettings `json:"settings" binding:"required"`
	Path 		string `json:"path" binding:"required"`
	Move 		DOMMatrixs `json:"move" binding:"required"`
	Pan 		DOMMatrixs `json:"pan" binding:"required"`
	LeftMost 	DOMMatrixs `json:"leftMost" binding:"required"`
	RightMost 	DOMMatrixs `json:"rightMost" binding:"required"`
	TopMost		DOMMatrixs `json:"topMost" binding:"required"`
	BottomMost 	DOMMatrixs `json:"bottomMost" binding:"required"`
}

type DOMMatrixs struct {
	A	int `json:"a" binding:"required"`
    B	int `json:"b" binding:"required"`
    C	int `json:"c" binding:"required"`
    D	int `json:"d" binding:"required"`
    E	int `json:"e" binding:"required"`
    F	int `json:"f" binding:"required"`
    M11	int `json:"m11" binding:"required"`
    M12	int `json:"m12" binding:"required"`
    M13	int `json:"m13" binding:"required"`
    M14	int `json:"m14" binding:"required"`
    M21	int `json:"m21" binding:"required"`
    M22	int `json:"m22" binding:"required"`
    M23	int `json:"m23" binding:"required"`
    M24	int `json:"m24" binding:"required"`
    M31	int `json:"m31" binding:"required"`
    M32	int `json:"m32" binding:"required"`
    M33	int `json:"m33" binding:"required"`
    M34	int `json:"m34" binding:"required"`
    M41	int `json:"m41" binding:"required"`
    M42	int `json:"m42" binding:"required"`
    M43	int `json:"m43" binding:"required"`
    M44	int `json:"m44" binding:"required"`
}
