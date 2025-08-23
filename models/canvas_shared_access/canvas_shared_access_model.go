package canvas_shared_access_model

import (
	model "qolboard-api/models"
	"qolboard-api/services/logging"

	"github.com/jmoiron/sqlx"
)

func GetAll(tx *sqlx.Tx, limit int, page int) ([]model.CanvasSharedAccess, error) {
	limit = min(limit, 100)
	offset := max(page-1, 0) * limit
	csa := make([]model.CanvasSharedAccess, 0)
	err := tx.Select(
		&csa,
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
		logging.LogError("[model]", "Error getting all canvases shared accesses", err)
		return nil, err
	}

	logging.Here()
	logging.LogDebug("[model]", "csa", csa)
	return csa, err
}
