package model

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
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
	UserUuid               string                    `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;index"`
	CanvasData             datatypes.JSON            `json:"canvas_data" db:"canvas_data"`
	CanvasSharedAccess     []*CanvasSharedAccess     `json:"canvas_shared_accesses" db:"canvas_shared_accesses"`
	CanvasSharedInvitation []*CanvasSharedInvitation `json:"canvas_shared_invitations" db:"canvas_shared_invitations"`
	User                   *User                     `json:"user" db:"user"`
}

type CanvasData struct {
	Name            string         `json:"name" binding:"required"`
	BackgroundColor string         `json:"backgroundColor" binding:"required"`
	PieceSettings   *PieceSettings `json:"pieceSettings" binding:"required"`
	RulerSettings   RulerSettings  `json:"rulerSettings"`
	PiecesManager   PiecesManager  `json:"piecesManager" binding:"required"`
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

func (c Canvas) Get(tx *sqlx.Tx, canvasId string) (*Canvas, error) {
	canvas := &Canvas{}
	err := tx.Get(canvas, "SELECT * FROM canvases c WHERE AND c.id = $1 AND deleted_at IS NULL", canvasId)
	if err != nil {
		return nil, err
	}

	return canvas, nil
}

func (c Canvas) GetAll(tx *sqlx.Tx) ([]Canvas, error) {
	var canvases []Canvas
	err := tx.Select(&canvases, "SELECT * FROM canvases c WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}

	return canvases, err
}

func (c *Canvas) Save(tx *sqlx.Tx) error {
	now := time.Now()
	canvasDataBytes, err := json.Marshal(c.CanvasData)
	if err != nil {
		return err
	}

	if c.ID > 0 {
		err = tx.Get(c, "UPDATE canvas SET canvas_data = $1, updated_at = $2 WHERE user_uuid = $3 AND id = $4 AND deleted_at IS NULL RETURNING *", string(canvasDataBytes), now, c.UserUuid, c.ID)
	} else {
		err = tx.Get(c, "INSERT INTO canvases(canvas_data, created_at, updated_at, user_uuid) VALUES($1, $2, $3, $4) RETURNING *", string(canvasDataBytes), now, now, c.UserUuid)
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *Canvas) Delete(tx *sqlx.Tx) error {
	now := time.Now()
	err := tx.Get(c, "UPDATE canvas SET deleted_at = $1 WHERE AND id = $2 AND deleted_at IS NULL RETURNING *", now, c.ID)

	return err
}

func (c Canvas) BelongsToUser(db *gorm.DB, userUuid string) *gorm.DB {
	return db.Where("canvas.user_uuid", userUuid)
}

func (c Canvas) LeftJoinCanvasSharedAccessOnUser(db *gorm.DB, userUuid string) *gorm.DB {
	return db.Joins("LEFT JOIN canvas_shared_accesses ON canvas_shared_accesses.user_uuid = ?", userUuid)
}
