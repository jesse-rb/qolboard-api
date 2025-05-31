package canvas_model

import (
	model "qolboard-api/models"

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

var relationLoaders model.RelationLoaders[model.Canvas] = model.RelationLoaders[model.Canvas]{
	BelongsTo: map[string]model.BelongsToLoader[model.Canvas]{
		"user": {
			Loader: func(tx *sqlx.Tx, c *model.Canvas) error {
				// user, err := model.User{}.Get(tx, c.UserUuid)
				user := &model.User{}
				err := tx.Select(user, "SELECT * FROM view_users u WHERE u.id = $1", c.UserUuid)
				if err != nil {
					return err
				}
				c.User = user
				return nil
			},
			BatchLoader: func(tx *sqlx.Tx, cSlice []model.Canvas) error {
				var with map[string]any = make(map[string]any, 0)

				// Get User uuids
				userUuids := []string{}
				for _, canvas := range cSlice {
					userUuids = append(userUuids, canvas.UserUuid)
				}

				// Get users by uuids
				users := []model.User{}
				query, args, err := sqlx.In("SELECT * FROM view_users u WHERE u.id IN (?);", userUuids)
				if err != nil {
					return err
				}

				query = tx.Rebind(query)
				err = tx.Select(&users, query, args...)
				if err != nil {
					return err
				}

				// Key users by uuid
				usersMap := map[string]*model.User{}
				for _, user := range users {
					usersMap[user.Uuid] = &user
				}

				// Mix in
				for i := range cSlice {
					cSlice[i].User = usersMap[cSlice[i].UserUuid]
				}

				return nil
			},
		},
	},
	HasOne: map[string]model.HasOneLoader[model.Canvas]{},
	HasMany: map[string]model.HasManyLoader[model.Canvas]{
		"canvas_shared_invitations": {
			Loader: func(tx *sqlx.Tx, c *model.Canvas) error {
				canvasSharedInvitations := make([]model.CanvasSharedInvitation, 0)
				err := tx.Select(canvasSharedInvitations, "SELECT * FROM canvas_shared_invitations csi WHERE csi.canvas_id = $1", c.ID)
				if err != nil {
					return err
				}

				c.CanvasSharedInvitations = canvasSharedInvitations

				return nil
			},
			BatchLoader: func(tx *sqlx.Tx, models []model.Canvas) error {
				// Get all canvas shared invitation ids
				canvasIds := make([]uint64, 0)
				for _, c := range models {
					canvasIds = append(canvasIds, c.ID)
				}

				// Get canvas shared invitations by canvas ids
				var csiSlice []model.CanvasSharedInvitation
				query, args, err := sqlx.In("SELECT * FROM canvas_shared_invitations csi WHERE csi.canvas_id IN (?);", canvasIds)
				if err != nil {
					return err
				}
				query = tx.Rebind(query)
				err = tx.Select(&csiSlice, query, args...)
				if err != nil {
					return err
				}

				// Key csiList by canvas id
				csiMap := make(map[uint64][]model.CanvasSharedInvitation, 0)
				for _, csi := range csiSlice {
					if _, ok := csiMap[csi.CanvasId]; !ok {
						csiMap[csi.CanvasId] = make([]model.CanvasSharedInvitation, 0)
					}

					csiMap[csi.CanvasId] = append(csiMap[csi.CanvasId], csi)
				}

				// Mix in
				for i := range models {
					if csiSlice, ok := csiMap[models[i].ID]; ok {
						models[i].CanvasSharedInvitations = csiSlice
					}
				}

				return nil
			},
		},
	},
}

func LoadRelations(canvas *model.Canvas, tx *sqlx.Tx, with []string) error {
	err := model.GenericRelationsLoader(relationLoaders, canvas, tx, with)
	return err
}

func LoadBatchRelations(canvases []model.Canvas, tx *sqlx.Tx, with map[string]any) error {
	err := model.GenericBatchRelationsLoader(relationLoaders, canvases, tx, with)
	return err
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
