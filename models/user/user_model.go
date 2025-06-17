package user_model

import (
	model "qolboard-api/models"
	relations_service "qolboard-api/services/relations"

	"github.com/jmoiron/sqlx"
)

var UserRelations = relations_service.NewRelationRegistry[model.User]()

func init() {
	// HasMany Canvases
	UserRelations.RegisterSingle("canvases", relations_service.MakeHasManySingleLoader(
		"SELECT * FROM canvases WHERE user_uuid = $1 AND deleted_at IS NULL",
		func(u *model.User) any { return u.Uuid },
		func(u *model.User, rels []model.Canvas) { u.Canvases = rels },
	))
	UserRelations.RegisterBatch("canvases", relations_service.MakeHasManyBatchLoader(
		"SELECT * FROM canvases WHERE user_uuid IN (?) AND deleted_at IS NULL",
		func(u *model.User) string { return u.Uuid },
		func(u *model.User, rels []model.Canvas) { u.Canvases = rels },
		func(c *model.Canvas) string { return c.UserUuid },
	))
}

func Get(tx *sqlx.Tx, userUuid string) (*model.User, error) {
	user := &model.User{}
	err := tx.Get(user, "SELECT * FROM view_users u WHERE u.id = $1", userUuid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
