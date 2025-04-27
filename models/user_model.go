package model

import (
	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Uuid  string `json:"uuid" gorm:"column:id;primaryKey;type:uuid" db:"id"`
	Email string `json:"email" db:"email"`
}

func (u User) TableName() string {
	return "auth.users"
}

func (u User) Get(tx *sqlx.Tx, userUuid string) (*User, error) {
	user := &User{}
	err := tx.Get(user, "SELECT * FROM view_users u WHERE u.id = $1", userUuid)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// func (u User) HasAccessToCanvas(tx *gorm.DB, userUuid string) *gorm.DB {
// 	tx = Canvas{}.LeftJoinCanvasSharedAccessOnUser(tx, userUuid)
// 	tx.Where(Canvas{}.BelongsToUser(tx, userUuid))
// 	tx = tx.Or(CanvasSharedAccess{}.BelongsToCanvas(tx, nil))
//
// 	return t
// }

func (u User) LoadRelations(with []string) {
}

func (u User) Response() map[string]any {
	r := imissphp.ToMap(u)
	return r
}
