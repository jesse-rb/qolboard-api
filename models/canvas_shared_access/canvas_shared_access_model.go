package canvas_shared_access

import (
	model "qolboard-api/models"
)

type CanvasSharedAccess struct {
	model.Model
	UserUuid string `json:"user_uuid" gorm:"not null"`
	CanvasId uint64 `json:"canvas_id" gorm:"not null"`
	CanvasSharedInvitationId uint64 `json:"canvas_shared_invitation_id" gorm:"not null"`
}