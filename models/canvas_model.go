package model

import (
	"encoding/json"
	relations_service "qolboard-api/services/relations"
	"time"

	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
	"gorm.io/datatypes"
)

type Canvas struct {
	Model
	UserUuid                string                   `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;index"`
	CanvasData              datatypes.JSON           `json:"canvas_data" db:"canvas_data"`
	CanvasSharedAccesses    []CanvasSharedAccess     `json:"canvas_shared_accesses" x_ismodel:"true"`
	CanvasSharedInvitations []CanvasSharedInvitation `json:"canvas_shared_invitations" x_ismodel:"true"`
	User                    *User                    `json:"user" x_ismodel:"true"`
}

var CanvasRelations relations_service.RelationRegistry = relations_service.NewRelationRegistry()

func (c Canvas) GetRelations() relations_service.RelationRegistry {
	return CanvasRelations
}

func (c Canvas) GetPrimaryKey() any {
	return c.ID
}

func init() {
	// Belongs to User
	relations_service.HasOne(
		"user",
		CanvasRelations,
		"SELECT * FROM view_users WHERE id = $1",
		"SELECT * FROM view_users WHERE id IN (?)",
		func(c *Canvas, u *User) { c.User = u },
		func(c Canvas) any { return c.UserUuid },
		func(u User) any { return u.Uuid },
	)

	// Has many CanvasSharedInvitations
	relations_service.HasMany(
		"canvas_shared_invitations",
		CanvasRelations,
		"SELECT * FROM canvas_shared_invitations WHERE id = $1",
		"SELECT * FROM canvas_shared_invitations WHERE id IN (?)",
		func(c *Canvas, csi []CanvasSharedInvitation) { c.CanvasSharedInvitations = csi },
		func(c Canvas) any { return c.ID },
		func(csi CanvasSharedInvitation) any { return csi.CanvasId },
	)
}

func (c *Canvas) Save(tx *sqlx.Tx) error {
	now := time.Now()
	canvasDataBytes, err := json.Marshal(c.CanvasData)
	if err != nil {
		return err
	}

	if c.ID > 0 {
		err = tx.Get(c, "UPDATE canvas SET canvas_data = $1, updated_at = $2 WHERE user_uuid = $3 AND id = $4 AND deleted_at IS NULL RETURNING *", string(canvasDataBytes), now, c.UserUuid, c.ID)
	} else {
		err = tx.Get(c, "INSERT INTO canvases(canvas_data, created_at, updated_at, user_uuid) VALUES($1, $2, $3, $4) RETURNING *", string(canvasDataBytes), now, now, c.UserUuid)
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *Canvas) Delete(tx *sqlx.Tx) error {
	now := time.Now()
	err := tx.Get(c, "UPDATE canvas SET deleted_at = $1 WHERE AND id = $2 AND deleted_at IS NULL RETURNING *", now, c.ID)

	return err
}

func (c Canvas) Response() map[string]any {
	// r := c.Model.Response()
	r := imissphp.ToMap(c)
	return r
}
