package model

import (
	"fmt"
	"os"
	service "qolboard-api/services"

	"gorm.io/gorm"
)

type CanvasSharedInvitation struct {
	Model
	Code               string              `json:"-" gorm:"not null;index:,unique"`
	CanvasId           uint64              `json:"canvas_id" gorm:"not null"`
	UserUuid           string              `json:"userUuid" gorm:"not null;index"`
	Canvas             *Canvas             `json:"canvas"`
	CanvasSharedAccess *CanvasSharedAccess `json:"canvas_shared_access"`

	InviteLink string `json:"link" gorm:"-"` // Calculated on the fly
}

func CanvasSharedInvitationBelongsToUser(userUuid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_uuid", userUuid)
	}
}

func CanvasSharedInvitationBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_id", canvasId)
	}
}

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

func (sharedInvitation *CanvasSharedInvitation) Response() *CanvasSharedInvitation {
	sharedInvitation.InviteLink = sharedInvitation.buildInviteLink()
	sharedInvitation.Code = ""
	return sharedInvitation
}

func (sharedInvitation *CanvasSharedInvitation) buildInviteLink() string {
	apiHost := os.Getenv("API_HOST")
	return fmt.Sprintf("%s/user/canvas/%v/accept_invite/%s", apiHost, sharedInvitation.CanvasId, sharedInvitation.Code)
}
