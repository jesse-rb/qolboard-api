package model

import (
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

func CanvasSharedAccessBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_shared_accesses.canvas_id", canvasId)
	}
}

func CanvasSharedAccessBelongsToCanvasSharedInvitation(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_shared_accesses.canvas_id", canvasId)
	}
}

func CanvasSharedAccessInnerJoinCanvasOnCanvasOwner(userUuid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("INNER JOIN canvas ON canvas.id = canvas_shared_accesses.canvas_id AND canvas.user_uuid = ?", userUuid)
	}
}

func CanvasSharedAccessLeftJoinCanvasOnCanvasOwner(userUuid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("LEFT JOIN canvas ON canvas.id = canvas_shared_accesses.canvas_id AND canvas.user_uuid = ?", userUuid)
	}
}

func CanvasSharedAccessBelongsToUser(userUuid string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_shared_accesses.user_uuid", userUuid)
	}
}
