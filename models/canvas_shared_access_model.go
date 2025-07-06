package model

import (
	relations_service "qolboard-api/services/relations"
	"time"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

type CanvasSharedAccess struct {
	Model
	DeletedAt                *gorm.DeletedAt         `gorm:"Index:uix_user_canvas" json:"deletedAt"` // Override to be included in our composite unique index
	UserUuid                 string                  `json:"user_uuid" gorm:"foreignKey:UserUuid;references:Uuid;type:uuid;not null;uniqueIndex:uix_user_canvas"`
	CanvasId                 uint64                  `json:"canvas_id" gorm:"not null;uniqueIndex:uix_user_canvas"`
	CanvasSharedInvitationId uint64                  `json:"canvas_shared_invitation_id" gorm:"not null"`
	Canvas                   *Canvas                 `json:"canvas"`
	CanvasSharedInvitation   *CanvasSharedInvitation `json:"canvas_shared_invitation"`
	User                     *User                   `json:"user"`
}

var CanvasSharedAccessRelations relations_service.RelationRegistry = relations_service.NewRelationRegistry()

func init() {
	relations_service.BelongsTo(
		"user",
		CanvasSharedAccessRelations,
		"SELECT * FROM view_users WHERE id = $1",
		"SELECT * FROM view_users WHERE id IN (?)",
		func(csa CanvasSharedAccess, u User) CanvasSharedAccess {
			csa.User = &u
			return csa
		},
		func(csa CanvasSharedAccess) any {
			return csa.UserUuid
		},
		func(u User) any {
			return u.Uuid
		},
	)
	relations_service.BelongsTo(
		"canvas",
		CanvasSharedAccessRelations,
		"SELECT * FROM canvases WHERE id = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvases WHERE id IN (?) AND deleted_at IS NULL",
		func(csa CanvasSharedAccess, c Canvas) CanvasSharedAccess {
			csa.Canvas = &c
			return csa
		},
		func(csa CanvasSharedAccess) any {
			return csa.CanvasId
		},
		func(c Canvas) any {
			return c.ID
		},
	)
	relations_service.BelongsTo(
		"canvas_shared_invitation",
		CanvasSharedAccessRelations,
		"SELECT * FROM canvas_shared_invitations WHERE id = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvas_shared_invitations WHERE id IN (?) AND deleted_at IS NULL",
		func(csa CanvasSharedAccess, csi CanvasSharedInvitation) CanvasSharedAccess {
			csa.CanvasSharedInvitation = &csi
			return csa
		},
		func(csa CanvasSharedAccess) any {
			return csa.CanvasSharedInvitationId
		},
		func(csi CanvasSharedInvitation) any {
			return csi.ID
		},
	)
}

func (csa CanvasSharedAccess) GetRelations() relations_service.RelationRegistry {
	return CanvasSharedAccessRelations
}

func (csa CanvasSharedAccess) GetPrimaryKey() any {
	return csa.ID
}

func (csa *CanvasSharedAccess) Insert(tx *sqlx.Tx) error {
	now := time.Now()

	err := tx.Get(csa, `
INSERT INTO canvas_shared_accesses(created_at, updated_at, user_uuid, canvas_id, canvas_shared_invitation_id)
VALUES($1, $2, get_user_uuid(), $3, $4) RETURNING *
	`, now, now, csa.CanvasId, csa.CanvasSharedInvitationId)
	if err != nil {
		return err
	}

	return nil
}

func (csa *CanvasSharedAccess) Delete(tx *sqlx.Tx) error {
	now := time.Now()

	err := tx.Get(csa, `
UPDATE canvas_shared_accesses
SET deleted_at = $1, updated_at = $2
WHERE id = $3
AND (
	user_uuid = get_user_uuid()
	OR (SELECT user_uuid FROM canvases WHERE canvas.id = canvas_shared_accesses.canvas_id) = get_user_uuid()
) RETURNING *
`,
		now, now, csa.ID,
	)
	return err
}
