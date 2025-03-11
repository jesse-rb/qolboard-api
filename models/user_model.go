package model

import (
	"gorm.io/gorm"
)

type User struct {
	Uuid  string `json:"uuid" gorm:"column:id;primaryKey;type:uuid"`
	Email string `json:"email"`
}

func (u User) TableName() string {
	return "auth.users"
}

func (u User) HasAccessToCanvas(tx *gorm.DB, userUuid string) *gorm.DB {
	tx = Canvas{}.LeftJoinCanvasSharedAccessOnUser(tx, userUuid)
	tx.Where(Canvas{}.BelongsToUser(tx, userUuid))
	tx = tx.Or(CanvasSharedAccess{}.BelongsToCanvas(tx, nil))

	return tx
}
