package canvas_shared_invitation_model

import (
	model "qolboard-api/models"
	service "qolboard-api/services"

	"github.com/jmoiron/sqlx"
)

func NewCanvasSharedInvitation(userId string, canvasId string) (*model.CanvasSharedInvitation, error) {
	code, err := service.GenerateCode(256)
	if err != nil {
		return nil, err
	}

	return &model.CanvasSharedInvitation{
		UserId:   userId,
		CanvasId: canvasId,
		Code:     code,
	}, nil
}

func GetByCode(tx *sqlx.Tx, canvasId string, code string) (model.CanvasSharedInvitation, error) {
	csi := model.CanvasSharedInvitation{}
	err := tx.Get(&csi, `
SELECT *
FROM canvas_shared_invitations csi
WHERE csi.canvas_id = $1
AND csi.code = $2
	`, canvasId, code)

	return csi, err
}

func GetAllForCanvas(tx *sqlx.Tx, canvasId string) ([]model.CanvasSharedInvitation, error) {
	var canvasSharedInvitiations []model.CanvasSharedInvitation
	err := tx.Select(&canvasSharedInvitiations, "SELECT * FROM canvas_shared_invitations csi WHERE canvas_id = $1 AND deleted_at IS NULL", canvasId)
	if err != nil {
		return nil, err
	}

	return canvasSharedInvitiations, err
}
