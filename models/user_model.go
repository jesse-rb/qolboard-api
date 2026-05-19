package model

import (
	"fmt"
	"qolboard-api/services/logging"
	relations_service "qolboard-api/services/relations"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	Uuid                     string               `json:"uuid" gorm:"column:id;primaryKey;type:uuid" db:"id"`
	Email                    string               `json:"email" db:"email"`
	EmailVerificationCode    *string              `json:"email_verification_code" db:"email_verification_code"`
	EmailVerificationCodeIAT *time.Time           `json:"email_verification_code_iat" db:"email_verification_code_iat"`
	LoginOTP                 *string              `json:"login_otp" db:"login_otp"`
	LoginOTPIAT              *time.Time           `json:"login_otp_iat" db:"login_otp_iat"`
	VerifiedAt               *time.Time           `json:"verified_at" db:"verified_at"`
	CreatedAt                string               `json:"created_at" db:"created_at"`
	UpdatedAt                string               `json:"updated_at" db:"updated_at"`
	DeletedAt                *string              `json:"deleted_at" db:"deleted_at"`
	Canvases                 []Canvas             `json:"canvases"`
	CanvasSharedAccesses     []CanvasSharedAccess `json:"canvas_shared_accesses"`
}

var UserRelations = relations_service.NewRelationRegistry()

func (u User) GetRelations() relations_service.RelationRegistry {
	return UserRelations
}

func (u User) GetPrimaryKey() any {
	return u.Uuid
}

func (u User) GetForeignKey(related relations_service.IHasRelations) any {
	fk := related.GetPrimaryKey()
	return fk
}

func init() {
	// HasMany Canvases
	relations_service.HasMany(
		"canvases",
		UserRelations,
		"SELECT * FROM canvases WHERE user_uuid = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvases WHERE user_uuid IN (?) AND deleted_at IS NULL",
		func(u User, c []Canvas) User { u.Canvases = c; return u },
		func(u User) any { return u.Uuid },
		func(c Canvas) any { return c.UserUuid },
	)

	relations_service.HasMany(
		"canvas_shared_accesses",
		UserRelations,
		"SELECT * FROM canvas_shared_accesses WHERE user_uuid = $1 AND deleted_at IS NULL",
		"SELECT * FROM canvas_shared_accesses WHERE user_uuid IN (?) AND deleted_at IS NULL",
		func(u User, csa []CanvasSharedAccess) User {
			u.CanvasSharedAccesses = csa
			return u
		},
		func(u User) any { return u.Uuid },
		func(csa CanvasSharedAccess) any {
			return csa.UserUuid
		},
	)
}

func (u User) Response() map[string]any {
	r := map[string]any{
		"uuid":                   u.Uuid,
		"email":                  u.Email,
		"canvases":               u.Canvases,
		"canvas_shared_accesses": u.CanvasSharedAccesses,
	}
	return r
}

func (u *User) Create(tx *sqlx.Tx) error {
	err := tx.Get(u, "INSERT INTO users(email, email_verification_code, email_verification_code_iat) VALUES($1, $2, $3) RETURNING *", u.Email, u.EmailVerificationCode, u.EmailVerificationCodeIAT)
	if err != nil {
		logging.LogError("[model]", "Error creating user", err)
		return err
	}

	return nil
}

func (u *User) Update(tx *sqlx.Tx, fieldsToUpdate []string) error {
	if u.Uuid == "" {
		return fmt.Errorf("user UUID not set")
	}
	if len(fieldsToUpdate) < 1 {
		return fmt.Errorf("failed to update user with no fields specified to update")
	}

	builder := strings.Builder{}
	params := make([]any, 0)
	builder.WriteString("UPDATE users SET")

	// Iterate over fields to update, to build update statement and params
	for i, fieldName := range fieldsToUpdate {
		if i > 0 {
			builder.WriteString(",")
		}
		switch fieldName {
		case "email_verification_code":
			params = append(params, u.EmailVerificationCode)
			fmt.Fprintf(&builder, " email_verification_code = $%d", i+1)
		case "email_verification_code_iat":
			params = append(params, u.EmailVerificationCodeIAT)
			fmt.Fprintf(&builder, " email_verification_code_iat = $%d", i+1)
		case "verified_at":
			params = append(params, u.VerifiedAt)
			fmt.Fprintf(&builder, " verified_at = $%d", i+1)
		default:
			return fmt.Errorf("failed to update user with unkown field specified to update")
		}
	}

	params = append(params, u.Uuid)
	fmt.Fprintf(&builder, " WHERE id = $%d RETURNING *;", len(params))

	err := tx.Get(u, builder.String(), params...)
	if err != nil {
		logging.LogError("[model]", "Error creating user", err)
		return err
	}

	return nil
}
