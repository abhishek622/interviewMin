package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository is the concrete implementation for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// Create inserts a new user and returns the new user's id.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (string, error) {
	id := uuid.New().String()
	const q = `
INSERT INTO users (user_id, email, password_hash, role, created_at, updated_at)
VALUES ($1, $2, $3, 'user', now(), now())
`
	_, err := r.db.Exec(ctx, q, id, email, passwordHash)
	if err != nil {
		// handle unique violation more gracefully if desired
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// PostgreSQL unique_violation code is "23505"
			if pgErr.Code == "23505" {
				return "", fmt.Errorf("email already exists: %w", err)
			}
		}
		return "", fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

// GetByEmail returns a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	const q = `
SELECT user_id, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1
`
	var u model.User
	row := r.db.QueryRow(ctx, q, email)
	if err := row.Scan(&u.UserID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("user not found: %w", err)
		}
		return model.User{}, fmt.Errorf("scan user by email: %w", err)
	}
	return u, nil
}

// GetByID returns a user by id.
func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	const q = `
SELECT user_id, email, password_hash, role, created_at, updated_at
FROM users
WHERE user_id = $1
`
	var u model.User
	row := r.db.QueryRow(ctx, q, id)
	if err := row.Scan(&u.UserID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("user not found: %w", err)
		}
		return model.User{}, fmt.Errorf("scan user by id: %w", err)
	}
	return u, nil
}
