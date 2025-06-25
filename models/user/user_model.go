package user_model

import (
	model "qolboard-api/models"

	"github.com/jmoiron/sqlx"
)

func Get(tx *sqlx.Tx, userUuid string) (*model.User, error) {
	user := &model.User{}
	err := tx.Get(user, "SELECT * FROM view_users u WHERE u.id = $1", userUuid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
