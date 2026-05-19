package database

import (
	"database/sql"
	"errors"
	"qolboard-api/services/logging"

	"github.com/jmoiron/sqlx"
)

func StandardDeferRollback(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		logging.LogError("auth controller", "failed defer tx rollback", err)
	}
}
