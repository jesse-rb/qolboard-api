package model

import (
	"time"
)

type CanvasSharedInvitation struct {
	Model
	Code      string    `json:"code" gorm:"not null"`
	CanvasId  uint64    `json:"canvas_id" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
}

