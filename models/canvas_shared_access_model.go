package model

import "gorm.io/gorm"

type CanvasSharedAccess struct {
	Model
	UserUuid                 string                  `json:"user_uuid" gorm:"not null"`
	CanvasId                 uint64                  `json:"canvas_id" gorm:"not null"`
	CanvasSharedInvitationId uint64                  `json:"canvas_shared_invitation_id" gorm:"not null"`
	Canvas                   *Canvas                 `json:"canvas"`
	CanvasSharedInvitation   *CanvasSharedInvitation `json:"canvas_shared_invitation"`
}

func CanvasSharedAccessBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_id", canvasId)
	}
}

func CanvasSharedAccessBelongsToCanvasSharedInvitation(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_id", canvasId)
	}
}
