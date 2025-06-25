package model

import (
	"fmt"
	"os"
	relations_service "qolboard-api/services/relations"
	"time"

	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
)

type CanvasSharedInvitation struct {
	Model
	Code               string                `json:"-" db:"code" gorm:"not null;index:,unique"`
	CanvasId           uint64                `json:"canvas_id" db:"canvas_id" gorm:"not null"`
	UserUuid           string                `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:id;type:uuid;not null;index"`
	Canvas             *Canvas               `json:"canvas"`
	CanvasSharedAccess []*CanvasSharedAccess `json:"canvas_shared_access"`

	InviteLink string `json:"link" gorm:"-"` // Calculated on the fly
}

var CanvasSharedInvitationRelations relations_service.RelationRegistry = relations_service.NewRelationRegistry()

func (csi CanvasSharedInvitation) GetRelations() relations_service.RelationRegistry {
	return CanvasSharedInvitationRelations
}

func (csi CanvasSharedInvitation) GetPrimaryKey() any {
	return csi.ID
}

func (csi *CanvasSharedInvitation) Save(tx *sqlx.Tx) error {
	now := time.Now()
	var err error

	if csi.ID > 0 {
		err = tx.Get(csi, "UPDATE canvas_shared_invitations SET updated_at = $1 WHERE user_uuid = $2 AND id = $3 AND deleted_at IS NULL RETURNING *", now, csi.UserUuid, csi.ID)
	} else {
		err = tx.Get(csi, "INSERT INTO canvas_shared_invitations(code, user_uuid, canvas_id, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING *", csi.Code, csi.UserUuid, csi.CanvasId, now, now)
	}

	if err != nil {
		return err
	}

	return nil
}

func (csi CanvasSharedInvitation) Response() map[string]any {
	csi.InviteLink = csi.buildInviteLink()
	r := imissphp.ToMap(csi)
	return r
}

func (sharedInvitation *CanvasSharedInvitation) buildInviteLink() string {
	apiHost := os.Getenv("API_HOST")
	return fmt.Sprintf("%s/user/canvas/%v/accept_invite/%s", apiHost, sharedInvitation.CanvasId, sharedInvitation.Code)
}
