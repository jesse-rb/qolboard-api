package model

import (
	"encoding/json"
	"fmt"
	service "qolboard-api/services"
	"qolboard-api/services/logging"
	relations_service "qolboard-api/services/relations"
	"time"

	"github.com/jmoiron/sqlx"
	"gorm.io/datatypes"
)

type Canvas struct {
	Model
	UserUuid                string                   `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;index"`
	CanvasData              datatypes.JSON           `json:"canvas_data" db:"canvas_data"`
	CanvasSharedAccesses    []CanvasSharedAccess     `json:"canvas_shared_accesses"`
	CanvasSharedInvitations []CanvasSharedInvitation `json:"canvas_shared_invitations"`
	User                    *User                    `json:"user"`
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
		func(c Canvas, u User) Canvas { c.User = &u; return c },
		func(c Canvas) any { return c.UserUuid },
		func(u User) any { return u.Uuid },
	)

	// Has many CanvasSharedInvitations
	relations_service.HasMany(
		"canvas_shared_invitations",
		CanvasRelations,
		"SELECT * FROM canvas_shared_invitations WHERE canvas_id = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvas_shared_invitations WHERE canvas_id IN (?) AND deleted_at IS NULL",
		func(c Canvas, csi []CanvasSharedInvitation) Canvas { c.CanvasSharedInvitations = csi; return c },
		func(c Canvas) any { return c.ID },
		func(csi CanvasSharedInvitation) any { return csi.CanvasId },
	)

	relations_service.HasMany(
		"canvas_shared_accesses",
		CanvasRelations,
		"SELECT * FROM canvas_shared_accesses WHERE canvas_id = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvas_shared_accesses WHERE canvas_id IN (?) AND deleted_at IS NULL",
		func(c Canvas, csa []CanvasSharedAccess) Canvas { c.CanvasSharedAccesses = csa; return c },
		func(c Canvas) any { return c.ID },
		func(csa CanvasSharedAccess) any { return csa.CanvasId },
	)
}

func (c *Canvas) Save(tx *sqlx.Tx) error {
	now := time.Now()
	canvasDataBytes, err := json.Marshal(c.CanvasData)
	if err != nil {
		return err
	}

	if c.ID > 0 {
		err = tx.Get(c, fmt.Sprintf(`
UPDATE canvases c
SET canvas_data = $1, updated_at = $2
WHERE %s
AND id = $3
AND deleted_at IS NULL
RETURNING *
		`, SqlHasAccessToCanvas("c")), string(canvasDataBytes), now, c.ID)
	} else {
		err = tx.Get(c, "INSERT INTO canvases(canvas_data, created_at, updated_at, user_uuid) VALUES($1, $2, $3, get_user_uuid()) RETURNING *", string(canvasDataBytes), now, now)
	}

	if err != nil {
		logging.LogError("[model]", "Error saving canvas", err)
		return err
	}

	return nil
}

func (c *Canvas) Delete(tx *sqlx.Tx) error {
	now := time.Now()

	err := tx.Get(c, "UPDATE canvases SET deleted_at = $1 WHERE id = $2 AND user_uuid = get_user_uuid() AND deleted_at IS NULL RETURNING *", now, c.ID)
	if err != nil {
		logging.LogError("[model]", "Error deleting canvas", err)
		return err
	}

	_, err = tx.Exec("UPDATE canvas_shared_invitations SET deleted_at = $1 WHERE canvas_id = $2 AND deleted_at IS NULL", now, c.ID)
	if err != nil {
		logging.LogError("[model]", "Error deleting related canvas shared invitations", err)
		return err
	}

	_, err = tx.Exec("UPDATE canvas_shared_accesses SET deleted_at = $1 WHERE canvas_id = $2 AND deleted_at IS NULL", now, c.ID)
	if err != nil {
		logging.LogError("[model]", "Error deleting related canvas shared access", err)
		return err
	}

	return err
}

func (c Canvas) Response() map[string]any {
	r := service.ToMapStringAny(c)
	return r
}

func SqlHasAccessToCanvas(aliasCanvas string) string {
	sql := fmt.Sprintf(`
(
	%s.user_uuid = get_user_uuid()
	OR EXISTS (
		SELECT csa.id
		FROM canvas_shared_accesses csa
		WHERE csa.user_uuid = get_user_uuid()
		AND csa.canvas_id = %s.id
		AND csa.deleted_at IS NULL
	)
)
	`, aliasCanvas, aliasCanvas)

	return sql
}
