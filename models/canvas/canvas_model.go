package canvas_model

import (
	"fmt"
	model "qolboard-api/models"
	"qolboard-api/services/logging"

	"github.com/jmoiron/sqlx"
)

func Get(tx *sqlx.Tx, canvasId uint64) (*model.Canvas, error) {
	canvas := &model.Canvas{}
	err := tx.Get(canvas, fmt.Sprintf(`
SELECT *
FROM canvases c
WHERE c.id = $1
AND deleted_at IS NULL
AND %s
	`, model.SqlHasAccessToCanvas("c")), canvasId)
	if err != nil {
		logging.LogError("[model]", "Error getting canvas", err)
		return nil, err
	}

	return canvas, nil
}

func GetAll(tx *sqlx.Tx, limit int, page int) ([]model.Canvas, error) {
	offset := max(page-1, 0) * limit
	limit = min(limit, 100)
	var canvases []model.Canvas
	err := tx.Select(&canvases, fmt.Sprintf(`
SELECT *
FROM canvases c
WHERE deleted_at IS NULL
AND %s
LIMIT $1
OFFSET $2
	`, model.SqlHasAccessToCanvas("c")), limit, offset)
	if err != nil {
		return nil, err
	}

	return canvases, err
}
