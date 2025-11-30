package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository is the concrete implementation for users.
type UserRepository struct {
	db *pgxpool.Pool
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
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
				return nil, fmt.Errorf("email already exists: %w", err)
			}
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
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

func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
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
