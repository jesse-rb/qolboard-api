package model

import (
	"fmt"
	relations_service "qolboard-api/services/relations"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type UserRefreshToken struct {
	ID           string     `json:"id" db:"id"`
	FamilyID     string     `json:"family_id" db:":family_id"`
	UserID       string     `json:"user_id" db:"user_id"`
	RefreshToken string     `json:"refresh_token" db:"refresh_token"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
	User         *User      `json:"user"`
}

var UserRefreshTokenRelations = relations_service.NewRelationRegistry()

func (urt UserRefreshToken) GetRelations() relations_service.RelationRegistry {
	return UserRelations
}

func (urt UserRefreshToken) GetPrimaryKey() any {
	return urt.ID
}

func (urt UserRefreshToken) GetForeignKey(related relations_service.IHasRelations) any {
	fk := related.GetPrimaryKey()
	return fk
}

func init() {
	// Has one user
	relations_service.HasOne(
		"user",
		UserRefreshTokenRelations,
		"SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL",
		"SELECT * FROM users WHERE id IN (?) AND deleted_at IS NULL",
		func(urt UserRefreshToken, u User) UserRefreshToken { urt.User = &u; return urt },
		func(urt UserRefreshToken) any { return urt.ID },
		func(u User) any { return u.Id },
	)
}

func (urt UserRefreshToken) Response() map[string]any {
	r := map[string]any{}
	return r
}

func (urt *UserRefreshToken) Create(tx *sqlx.Tx) error {
	stmt := strings.Builder{}

	paramNames := []string{
		"user_id",
		"refresh_token",
	}
	paramValues := []any{
		urt.UserID,
		urt.RefreshToken,
	}

	if urt.FamilyID != "" {
		paramNames = append(paramNames, "family_id")
		paramValues = append(paramValues, urt.FamilyID)
	}

	paramInserts := []string{}
	for i := range paramValues {
		paramInserts = append(paramInserts, fmt.Sprintf("$%d", i+1))
	}

	_, err := fmt.Fprintf(&stmt, "INSERT INTO user_refresh_tokens(%s) VALUES(%s)", strings.Join(paramNames, ", "), strings.Join(paramInserts, ", "))
	if err != nil {
		return fmt.Errorf("failed to write string: %w", err)
	}

	err = tx.Get(urt, stmt.String(), paramValues...)
	if err != nil {
		return fmt.Errorf("failed to create user refresh token: %w", err)
	}

	return nil
}

func (urt *UserRefreshToken) DeleteByRefreshToken(tx *sqlx.Tx) error {
	err := tx.Get(urt, "UPDATE user_refresh_tokens SET deleted_at = NOW() WHERE refresh_token = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING *", urt.RefreshToken, urt.UserID)
	if err != nil {
		return fmt.Errorf("failed to create user refresh token: %w", err)
	}

	return nil
}

func (urt *UserRefreshToken) DeleteByFamilyID(tx *sqlx.Tx) error {
	err := tx.Get(urt, "UPDATE user_refresh_tokens SET deleted_at = NOW() WHERE family_id = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING *", urt.FamilyID, urt.UserID)
	if err != nil {
		return fmt.Errorf("failed to create user refresh token: %w", err)
	}

	return nil
}

func (urt *UserRefreshToken) FindByRefreshToken(tx *sqlx.Tx) error {
	err := tx.Get(urt, "SELECT * FROM user_refresh_tokens WHERE refresh_token = $1 AND user_id = $2 NULL RETURNING *")
	if err != nil {
		return fmt.Errorf("failed to find user refresh token: %w", err)
	}
	return nil
}
