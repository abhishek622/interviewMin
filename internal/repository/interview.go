package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InterviewRepository is the concrete implementation for interviews.
type InterviewRepository struct {
	db *pgxpool.Pool
}

// Create inserts a new interview and returns its id
func (r *InterviewRepository) Create(ctx context.Context, userID, sourceID, title, rawText string) (string, error) {
	id := uuid.New().String()
	const q = `
INSERT INTO interviews (id, user_id, source_id, title, raw_text, created_at)
VALUES ($1, $2, $3, $4, $5, now())
`
	_, err := r.db.Exec(ctx, q, id, userID, sourceID, title, rawText)
	if err != nil {
		return "", fmt.Errorf("insert interview: %w", err)
	}
	return id, nil
}

// GetByID fetches an interview by id
func (r *InterviewRepository) GetByID(ctx context.Context, id string) (model.Interview, error) {
	const q = `
SELECT id, user_id, source_id, title, raw_text, created_at
FROM interviews
WHERE id = $1
`
	var it model.Interview
	row := r.db.QueryRow(ctx, q, id)
	err := row.Scan(&it.ID, &it.UserID, &it.SourceID, &it.Title, &it.RawText, &it.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Interview{}, fmt.Errorf("interview not found: %w", err)
		}
		return model.Interview{}, fmt.Errorf("scan interview: %w", err)
	}
	return it, nil
}

// ListByUser returns interviews for a user with pagination and total count
func (r *InterviewRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.Interview, int, error) {
	// total count
	var total int
	const countQ = `SELECT COUNT(1) FROM interviews WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count interviews: %w", err)
	}

	const q = `
SELECT id, user_id, source_id, title, raw_text, created_at
FROM interviews
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query interviews: %w", err)
	}
	defer rows.Close()

	out := make([]model.Interview, 0, 8)
	for rows.Next() {
		var it model.Interview
		if err := rows.Scan(&it.ID, &it.UserID, &it.SourceID, &it.Title, &it.RawText, &it.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan interview row: %w", err)
		}
		out = append(out, it)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("rows error: %w", rows.Err())
	}
	return out, total, nil
}
