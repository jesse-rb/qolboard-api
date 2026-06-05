package user_model

import (
	"fmt"
	model "qolboard-api/models"
	"time"

	"github.com/jmoiron/sqlx"
)

func Get(tx *sqlx.Tx) (*model.User, error) {
	user := &model.User{}
	err := tx.Get(user, "SELECT * FROM users u WHERE u.id = get_user_uuid()")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetByEmail(tx *sqlx.Tx, email string) (*model.User, error) {
	user := &model.User{}
	err := tx.Get(user, "SELECT * FROM users u WHERE u.email = $1", email)
	if err != nil {
		return nil, fmt.Errorf("error querying user by email: %w", err)
	}

	return user, nil
}

func GetByEmailVerificationCode(tx *sqlx.Tx, emailVerificationCode string) (*model.User, error) {
	user := &model.User{}
	expiredThreshold := time.Now().Add(-1 * time.Hour)

	err := tx.Get(user, "SELECT * FROM users u WHERE u.email_verification_code = $1 AND u.verified_at IS NULL AND u.email_verification_code_iat > $2", emailVerificationCode, expiredThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email verification code: %w", err)
	}

	return user, nil
}
