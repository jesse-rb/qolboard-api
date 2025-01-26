package model

import (
	"fmt"
	"os"
	service "qolboard-api/services"

	"gorm.io/gorm"
)

type CanvasSharedInvitation struct {
	Model
	Code               string              `json:"code" gorm:"not null;index:,unique"`
	CanvasId           uint64              `json:"canvas_id" gorm:"not null"`
	UserUuid           string              `json:"userUuid" gorm:"not null"`
	Canvas             *Canvas             `json:"canvas"`
	CanvasSharedAccess *CanvasSharedAccess `json:"canvas_shared_access"`

	InviteLink string `json:"link" gorm:"-"` // Calculated on the fly
}

func CanvasSharedInvitationBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_id", canvasId)
	}
}

func NewCanvasSharedInvitation(canvasId uint64) (*CanvasSharedInvitation, error) {
	code, err := service.GenerateCode(256)
	if err != nil {
		return nil, err
	}

	return &CanvasSharedInvitation{
		CanvasId: canvasId,
		Code:     code,
	}, nil
}

func (self *CanvasSharedInvitation) Response() *CanvasSharedInvitation {
	self.InviteLink = self.buildInviteLink()
	return self
}

func (self *CanvasSharedInvitation) buildInviteLink() string {
	apiHost := os.Getenv("API_HOST")
	return fmt.Sprintf("%s/canvas_shared_invite/%s", apiHost, self.Code)
}
