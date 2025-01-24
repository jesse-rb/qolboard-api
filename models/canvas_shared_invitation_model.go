package model

import (
	"crypto/rand"
	"encoding/base64"
	"qolboard-api/services/logging"
	"time"

	"gorm.io/gorm"
)

type CanvasSharedInvitation struct {
	Model
	Code               string              `json:"code" gorm:"not null"`
	CanvasId           uint64              `json:"canvas_id" gorm:"not null"`
	ExpiresAt          time.Time           `json:"expires_at" gorm:"not null"`
	Canvas             *Canvas             `json:"canvas"`
	CanvasSharedAccess *CanvasSharedAccess `json:"canvas_shared_access"`
}

func CanvasSharedInvitationBelongsToCanvas(canvasId uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("canvas_id", canvasId)
	}
}

func New(canvasId uint64) (*CanvasSharedInvitation, error) {
	code, err := generateCode(244)
	if err != nil {
		return nil, err
	}

	return &CanvasSharedInvitation{
		CanvasId: canvasId,
		Code:     code,
	}, nil
}

func generateCode(len uint) (string, error) {
	// Number of bytes needed for len base64 encoded chars
	lenBytes := (len*6 + 7) / 8

	// Init byte slice
	randomBytes := make([]byte, lenBytes)

	// Get random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		logging.LogError("generateCode", "Error generatiing code", err)
		return "", err
	}

	// Encode bytes to URL-safe Base64 string
	code := base64.RawURLEncoding.EncodeToString(randomBytes)

	return code, nil
}
