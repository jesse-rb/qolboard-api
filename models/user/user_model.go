package user_model

import (
	model "qolboard-api/models"

	"github.com/jmoiron/sqlx"
)

var relationLoaders model.RelationLoaders[model.User] = model.RelationLoaders[model.User]{
	BelongsTo: map[string]model.BelongsToLoader[model.User]{},
	HasOne:    map[string]model.HasOneLoader[model.User]{},
	HasMany: map[string]model.HasManyLoader[model.User]{
		"canvases": {
			Loader: func(tx *sqlx.Tx, user *model.User) error {
				canvases := make([]model.Canvas, 0)
				err := tx.Select(&canvases, "SELECT * FROM canvases c WHERE c.user_uuid = $1", user.Uuid)
				if err != nil {
					return err
				}

				user.Canvases = canvases

				return nil
			},
			BatchLoader: func(tx *sqlx.Tx, users []model.User) error {
				// Get user uuids
				userUuids := make([]string, 0)
				for _, u := range users {
					userUuids = append(userUuids, u.Uuid)
				}

				// Get all canvases by user Uuids
				canvasSlice := make([]model.Canvas, 0)
				query, args, err := sqlx.In("SELECT * FROM canvases c WHERE c.user_uuid IN (?)", userUuids)
				if err != nil {
					return err
				}

				query = tx.Rebind(query)
				err = tx.Select(&canvasSlice, query, args)
				if err != nil {
					return err
				}

				// Key canvases by user uuid
				canvasMap := make(map[string][]model.Canvas, 0)
				for _, c := range canvasSlice {
					if _, ok := canvasMap[c.UserUuid]; !ok {
						canvasMap[c.UserUuid] = make([]model.Canvas, 0)
					}

					canvasMap[c.UserUuid] = append(canvasMap[c.UserUuid], c)
				}

				// Mixin
				for i := range users {
					if canvasSlice, ok := canvasMap[users[i].Uuid]; ok {
						users[i].Canvases = canvasSlice
					}
				}

				return nil
			},
		},
	},
}

func LoadRelations(tx *sqlx.Tx, user *model.User, with []string) error {
	return model.GenericRelationsLoader(relationLoaders, user, tx, with)
}

func LoadBatchRelations(tx *sqlx.Tx, users []model.User, with []string) error {
	return model.GenericBatchRelationsLoader(relationLoaders, users, tx, with)
}

func Get(tx *sqlx.Tx, userUuid string) (*model.User, error) {
	user := &model.User{}
	err := tx.Get(user, "SELECT * FROM view_users u WHERE u.id = $1", userUuid)
	if err != nil {
		return nil, err
	}

	return user, nil
}
