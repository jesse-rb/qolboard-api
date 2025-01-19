package model

type CanvasSharedAccess struct {
	Model
	UserUuid                 string `json:"user_uuid" gorm:"not null"`
	CanvasId                 uint64 `json:"canvas_id" gorm:"not null"`
	CanvasSharedInvitationId uint64 `json:"canvas_shared_invitation_id" gorm:"not null"`
}

