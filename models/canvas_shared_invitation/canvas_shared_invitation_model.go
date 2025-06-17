package canvas_shared_invitation_model

import (
	model "qolboard-api/models"
	service "qolboard-api/services"

	"github.com/jmoiron/sqlx"
)

// var relationLoaders model.RelationLoaders[model.CanvasSharedInvitation] = model.RelationLoaders[model.CanvasSharedInvitation]{
// 	BelongsTo: map[string]model.BelongsToLoader[model.CanvasSharedInvitation]{
// 		"canvas": {
// 			Loader: func(tx *sqlx.Tx, m *model.CanvasSharedInvitation) error {
// 				return nil
// 			},
// 			BatchLoader: func(tx *sqlx.Tx, m []model.CanvasSharedInvitation) error {
// 				return nil
// 			},
// 		},
// 	},
// 	HasOne:  map[string]model.HasOneLoader[model.CanvasSharedInvitation]{},
// 	HasMany: map[string]model.HasManyLoader[model.CanvasSharedInvitation]{},
// }

func NewCanvasSharedInvitation(userUuid string, canvasId uint64) (*model.CanvasSharedInvitation, error) {
	code, err := service.GenerateCode(256)
	if err != nil {
		return nil, err
	}

	return &model.CanvasSharedInvitation{
		UserUuid: userUuid,
		CanvasId: canvasId,
		Code:     code,
	}, nil
}

func GetAllForCanvas(tx *sqlx.Tx, canvasId uint64) ([]model.CanvasSharedInvitation, error) {
	var canvasSharedInvitiations []model.CanvasSharedInvitation
	err := tx.Select(&canvasSharedInvitiations, "SELECT * FROM canvas_shared_invitations csi WHERE canvas_id = $1 AND deleted_at IS NULL", canvasId)
	if err != nil {
		return nil, err
	}

	return canvasSharedInvitiations, err
}
