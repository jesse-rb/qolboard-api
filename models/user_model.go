package model

import (
	relations_service "qolboard-api/services/relations"

	"github.com/jesse-rb/imissphp-go"
)

type User struct {
	Uuid                 string               `json:"uuid" gorm:"column:id;primaryKey;type:uuid" db:"id"`
	Email                string               `json:"email" db:"email"`
	Canvases             []Canvas             `json:"canvases"`
	CanvasSharedAccesses []CanvasSharedAccess `json:"canvas_shared_accesses"`
}

var UserRelations = relations_service.NewRelationRegistry()

func (u User) GetRelations() relations_service.RelationRegistry {
	return UserRelations
}

func (u User) GetPrimaryKey() any {
	return u.Uuid
}

func (u User) GetForeignKey(related relations_service.IHasRelations) any {
	fk := related.GetPrimaryKey()
	return fk
}

func init() {
	// HasMany Canvases
	relations_service.HasMany(
		"canvases",
		UserRelations,
		"SELECT * FROM canvases WHERE user_uuid = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvases WHERE user_uuid IN (?) AND deleted_at IS NULL",
		func(u User, c []Canvas) User { u.Canvases = c; return u },
		func(u User) any { return u.Uuid },
		func(c Canvas) any { return c.UserUuid },
	)

	relations_service.HasMany(
		"canvas_shared_accesses",
		UserRelations,
		"SELECT * FROM canvas_shared_accesses WHERE user_uuid = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvas_shared_accesses WHERE user_uuid IN (?) AND deleted_at IS NULL",
		func(u User, csa []CanvasSharedAccess) User {
			u.CanvasSharedAccesses = csa
			return u
		},
		func(u User) any { return u.Uuid },
		func(csa CanvasSharedAccess) any {
			return csa.UserUuid
		},
	)
}

func (u User) Response() map[string]any {
	r := imissphp.ToMap(u)
	return r
}
