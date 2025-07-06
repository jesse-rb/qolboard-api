package canvas_shared_access_model

import (
	model "qolboard-api/models"

	"github.com/jmoiron/sqlx"
)

func GetAll(tx *sqlx.Tx, limit int, page int) ([]model.CanvasSharedAccess, error) {
	limit = min(limit, 100)
	offset := max(page-1, 0) * limit
	csa := make([]model.CanvasSharedAccess, 0)
	err := tx.Select(
		tx,
		`
SELECT *
FROM canvas_shared_accesses
WHERE user_uuid = get_user_uuid()
AND deleted_at IS NULL
LIMIT $1
OFFSET $2
		`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return csa, err
}
