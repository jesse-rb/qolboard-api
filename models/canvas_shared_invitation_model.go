package model

import (
	"fmt"
	"os"
	service "qolboard-api/services"
	"time"

	"github.com/jmoiron/sqlx"
)

type CanvasSharedInvitation struct {
	Model
	Code               string              `json:"-" db:"code" gorm:"not null;index:,unique"`
	CanvasId           uint64              `json:"canvas_id" db:"canvas_id" gorm:"not null"`
	UserUuid           string              `json:"user_uuid" db:"user_uuid" gorm:"foreignKey:UserUuid;references:id;type:uuid;not null;index"`
	Canvas             *Canvas             `json:"canvas" db:"canvas"`
	CanvasSharedAccess *CanvasSharedAccess `json:"canvas_shared_access" db:"canvas_shared_access"`

	InviteLink string `json:"link" gorm:"-"` // Calculated on the fly
}

// func CanvasSharedInvitationBelongsToUser(userUuid string) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		return db.Where("user_uuid", userUuid)
// 	}
// }
//
// func CanvasSharedInvitationBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		return db.Where("canvas_id", canvasId)
// 	}
// }

func NewCanvasSharedInvitation(userUuid string, canvasId uint64) (*CanvasSharedInvitation, error) {
	code, err := service.GenerateCode(256)
	if err != nil {
		return nil, err
	}

	return &CanvasSharedInvitation{
		UserUuid: userUuid,
		CanvasId: canvasId,
		Code:     code,
	}, nil
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

func (sharedInvitation *CanvasSharedInvitation) Response() *CanvasSharedInvitation {
	sharedInvitation.InviteLink = sharedInvitation.buildInviteLink()
	sharedInvitation.Code = ""
	return sharedInvitation
}

func (sharedInvitation *CanvasSharedInvitation) buildInviteLink() string {
	apiHost := os.Getenv("API_HOST")
	return fmt.Sprintf("%s/user/canvas/%v/accept_invite/%s", apiHost, sharedInvitation.CanvasId, sharedInvitation.Code)
}
