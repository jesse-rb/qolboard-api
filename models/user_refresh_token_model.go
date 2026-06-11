package model

import (
	"fmt"
	relations_service "qolboard-api/services/relations"

	"github.com/jmoiron/sqlx"
)

type UserRefreshToken struct {
	ID           string  `json:"id" db:"id"`
	UserID       string  `json:"user_id" db:"user_id"`
	RefreshToken *string `json:"refresh_token" db:"refresh_token"`
	CreatedAt    string  `json:"created_at" db:"created_at"`
	User         User    `json:"user"`
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
		func(urt UserRefreshToken, u User) UserRefreshToken { urt.User = u; return urt },
		func(urt UserRefreshToken) any { return urt.ID },
		func(u User) any { return u.Id },
	)
}

func (urt UserRefreshToken) Response() map[string]any {
	r := map[string]any{}
	return r
}

func (urt *UserRefreshToken) Create(tx *sqlx.Tx) error {
	err := tx.Get(urt, "INSERT INTO user_refresh_tokens(user_id, refresh_token, created_at) VALUES($1, $2, $3) RETURNING *", urt.UserID, urt.RefreshToken, urt.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user refresh token: %w", err)
	}

	return nil
}

func (urt *UserRefreshToken) Delete(tx *sqlx.Tx) error {
	err := tx.Get(urt, "UPDATE user_refresh_tokens SET deleted_at = NOW() WHERE id = $2 RETURNING *", urt.ID)
	if err != nil {
		return fmt.Errorf("failed to create user refresh token: %w", err)
	}

	return nil
}
