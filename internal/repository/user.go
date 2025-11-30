package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) CreateUser(ctx context.Context, u *model.User) error {
	const q = `
	INSERT INTO users (name, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING user_id, is_admin, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, q, u.Name, u.Email, u.PasswordHash)
	if err := row.Scan(&u.UserID, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// PostgreSQL unique_violation code is "23505"
			if pgErr.Code == "23505" {
				return fmt.Errorf("email already exists: %w", err)
			}
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	const q = `
SELECT user_id, name, email, password_hash, is_admin, created_at, updated_at
FROM users
WHERE email = $1
`
	var u model.User
	row := r.db.QueryRow(ctx, q, email)
	if err := row.Scan(&u.UserID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("user not found: %w", err)
		}
		return model.User{}, fmt.Errorf("scan user by email: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (model.User, error) {
	const q = `
SELECT user_id, name, email, password_hash, is_admin, created_at, updated_at
FROM users WHERE user_id = $1
`
	var u model.User
	row := r.db.QueryRow(ctx, q, id)
	if err := row.Scan(&u.UserID, &u.Name, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("user not found: %w", err)
		}
		return model.User{}, fmt.Errorf("scan user by id: %w", err)
	}
	return u, nil
}

func (r *Repository) CreateUserSession(ctx context.Context, session *model.UserToken) (*model.UserToken, error) {
	const q = `
INSERT INTO user_tokens (user_token_id, user_id, refresh_token, expires_at, device_info, is_revoked)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING user_token_id
`
	row := r.db.QueryRow(ctx, q, session.UserTokenID, session.UserID, session.RefreshToken, session.ExpiresAt, session.DeviceInfo, session.IsRevoked)
	if err := row.Scan(&session.UserTokenID); err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	return session, nil
}

func (r *Repository) GetUserSession(ctx context.Context, userTokenId string) (*model.UserToken, error) {
	const q = `
SELECT user_token_id, user_id, refresh_token, expires_at, device_info, is_revoked, created_at
FROM user_tokens WHERE user_token_id = $1
`
	var session model.UserToken
	row := r.db.QueryRow(ctx, q, userTokenId)
	if err := row.Scan(&session.UserTokenID, &session.UserID, &session.RefreshToken, &session.ExpiresAt, &session.DeviceInfo, &session.IsRevoked, &session.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("session not found: %w", err)
		}
		return nil, fmt.Errorf("scan session: %w", err)
	}
	return &session, nil
}

func (r *Repository) RevokeUserSession(ctx context.Context, userTokenId string) error {
	const q = `UPDATE user_tokens SET is_revoked = true WHERE user_token_id = $1`
	_, err := r.db.Exec(ctx, q, userTokenId)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *Repository) DeleteUserSession(ctx context.Context, refreshToken string) error {
	const q = `DELETE FROM user_tokens WHERE refresh_token = $1`
	_, err := r.db.Exec(ctx, q, refreshToken)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
