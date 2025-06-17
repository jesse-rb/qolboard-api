package canvas_model

import (
	model "qolboard-api/models"
	relations_service "qolboard-api/services/relations"

	"github.com/jmoiron/sqlx"
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

var CanvasRelations = relations_service.NewRelationRegistry[model.Canvas]()

func init() {
	// Belongs to User
	CanvasRelations.RegisterSingle("user", relations_service.MakeSingleLoader(
		"SELECT * FROM view_users WHERE id = $1",
		func(c *model.Canvas) any { return c.UserUuid },
		func(c *model.Canvas, u *model.User) { c.User = u },
	))
	CanvasRelations.RegisterBatch("user", relations_service.MakeBatchLoader(
		"SELECT * FROM view_users WHERE id IN (?)",
		func(c *model.Canvas) string { return c.UserUuid },
		func(c *model.Canvas, u *model.User) { c.User = u },
		func(u *model.User) string { return u.Uuid },
	))

	// HasMany CanvasSharedIvnitations
	CanvasRelations.RegisterSingle("canvas_shared_invitations", relations_service.MakeHasManySingleLoader(
		"SELECT * FROM canvas_shared_invitations WHERE canvas_id = $1",
		func(c *model.Canvas) any { return c.ID },
		func(c *model.Canvas, rels []model.CanvasSharedInvitation) { c.CanvasSharedInvitations = rels },
	))
	CanvasRelations.RegisterBatch("canvas_shared_invitations", relations_service.MakeHasManyBatchLoader(
		"SELECT * FROM canvas_shared_invitations WHERE canvas_id IN (?)",
		func(c *model.Canvas) uint64 { return c.ID },
		func(c *model.Canvas, invs []model.CanvasSharedInvitation) { c.CanvasSharedInvitations = invs },
		func(i *model.CanvasSharedInvitation) uint64 { return i.CanvasId },
	))
}

func Get(tx *sqlx.Tx, canvasId uint64) (*model.Canvas, error) {
	canvas := &model.Canvas{}
	err := tx.Get(canvas, "SELECT * FROM canvases c WHERE c.id = $1 AND deleted_at IS NULL", canvasId)
	if err != nil {
		return nil, err
	}

	return canvas, nil
}

func GetAll(tx *sqlx.Tx, limit int, page int) ([]model.Canvas, error) {
	offset := max(page-1, 0) * limit
	var canvases []model.Canvas
	err := tx.Select(&canvases, "SELECT * FROM canvases c WHERE deleted_at IS NULL LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}

	return canvases, err
}
