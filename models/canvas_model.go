package model

import (
	"encoding/json"
	"time"

	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
	"gorm.io/datatypes"
)

const (
	CanvasModeDraw   = "draw"
	CanvasModeGrab   = "grab"
	CanvasModePan    = "pan"
	CanvasModeRemove = "remove"
)

type Canvas struct {
	Model
	UserUuid                string                   `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;index"`
	CanvasData              datatypes.JSON           `json:"canvas_data" db:"canvas_data"`
	CanvasSharedAccesses    []CanvasSharedAccess     `json:"canvas_shared_accesses" x_ismodel:"true"`
	CanvasSharedInvitations []CanvasSharedInvitation `json:"canvas_shared_invitations" x_ismodel:"true"`
	User                    *User                    `json:"user" x_ismodel:"true"`
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

var relationLoaders RelationLoaders[Canvas] = RelationLoaders[Canvas]{
	BelongsTo: map[string]BelongsToLoader[Canvas]{
		"user": {
			Loader: func(tx *sqlx.Tx, model Canvas) error {
				user, err := User{}.Get(tx, model.UserUuid)
				if err != nil {
					return err
				}
				model.User = user
				return nil
			},
			BatchLoader: func(tx *sqlx.Tx, models []Canvas) error {
				// Get User uuids
				userUuids := []string{}
				for _, canvas := range models {
					userUuids = append(userUuids, canvas.UserUuid)
				}

				// Get users by uuids
				users := []User{}
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
				usersMap := map[string]*User{}
				for _, user := range users {
					usersMap[user.Uuid] = &user
				}

				// Mix in
				for i := range models {
					models[i].User = usersMap[models[i].UserUuid]
				}

				return nil
			},
		},
	},
	HasMany: map[string]HasManyLoader[Canvas]{
		"canvas_shared_invitations": {
			Loader: func(tx *sqlx.Tx, model Canvas) error {
				canvasSharedInvitations, err := CanvasSharedInvitation{}.GetAllForCanvas(tx, model.ID)
				if err != nil {
					return err
				}

				model.CanvasSharedInvitations = canvasSharedInvitations

				return nil
			},
			BatchLoader: func(tx *sqlx.Tx, models []Canvas) error {
				// Get all canvas shared invitation ids
				canvasIds := make([]uint64, 0)
				for _, c := range models {
					canvasIds = append(canvasIds, c.ID)
				}

				// Get canvas shared invitations by canvas ids
				var csiSlice []CanvasSharedInvitation
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
				csiMap := make(map[uint64][]CanvasSharedInvitation, 0)
				for _, csi := range csiSlice {
					if _, ok := csiMap[csi.CanvasId]; !ok {
						csiMap[csi.CanvasId] = make([]CanvasSharedInvitation, 0)
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

func (c Canvas) LoadRelations(tx *sqlx.Tx, with []string) error {
	err := genericRelationsLoader(relationLoaders, c, tx, with)
	return err
}

func LoadBatchRelations(canvases []Canvas, tx *sqlx.Tx, with []string) error {
	err := genericBatchRelationsLoader(relationLoaders, canvases, tx, with)
	return err
}

func (c Canvas) Get(tx *sqlx.Tx, canvasId uint64) (*Canvas, error) {
	canvas := &Canvas{}
	err := tx.Get(canvas, "SELECT * FROM canvases c WHERE c.id = $1 AND deleted_at IS NULL", canvasId)
	if err != nil {
		return nil, err
	}

	return canvas, nil
}

func (c Canvas) GetAll(tx *sqlx.Tx, limit int, page int, with []string) ([]Canvas, error) {
	offset := max(page-1, 0) * limit
	var canvases []Canvas
	err := tx.Select(&canvases, "SELECT * FROM canvases c WHERE deleted_at IS NULL LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}

	err = LoadBatchRelations(canvases, tx, with)
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

// func (c Canvas) BelongsToUser(db *gorm.DB, userUuid string) *gorm.DB {
// 	return db.Where("canvas.user_uuid", userUuid)
// }
//
// func (c Canvas) LeftJoinCanvasSharedAccessOnUser(db *gorm.DB, userUuid string) *gorm.DB {
// 	return db.Joins("LEFT JOIN canvas_shared_accesses ON canvas_shared_accesses.user_uuid = ?", userUuid)
// }

func (c Canvas) Response() map[string]any {
	// r := c.Model.Response()
	r := imissphp.ToMap(c)
	return r
}
