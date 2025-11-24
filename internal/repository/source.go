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

// SourceRepository is the concrete implementation for sources.
type SourceRepository struct {
	db *pgxpool.Pool
}

// Create inserts a new source (e.g., leetcode, reddit) and returns its id.
func (r *SourceRepository) Create(ctx context.Context, name string) (string, error) {
	id := uuid.New().String()
	const q = `
INSERT INTO sources (id, name, created_at)
VALUES ($1, $2, now())
`
	_, err := r.db.Exec(ctx, q, id, name)
	if err != nil {
		// handle unique violation if you have unique constraint on name
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", fmt.Errorf("source already exists: %w", err)
		}
		return "", fmt.Errorf("insert source: %w", err)
	}
	return id, nil
}

// GetByID returns a source by ID.
func (r *SourceRepository) GetByID(ctx context.Context, id string) (model.Source, error) {
	const q = `
SELECT id, name, created_at
FROM sources
WHERE id = $1
`
	var s model.Source
	row := r.db.QueryRow(ctx, q, id)
	if err := row.Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Source{}, fmt.Errorf("source not found: %w", err)
		}
		return model.Source{}, fmt.Errorf("scan source by id: %w", err)
	}
	return s, nil
}

// GetByName returns a source by name.
func (r *SourceRepository) GetByName(ctx context.Context, name string) (model.Source, error) {
	const q = `
SELECT id, name, created_at
FROM sources
WHERE lower(name) = lower($1)
LIMIT 1
`
	var s model.Source
	row := r.db.QueryRow(ctx, q, name)
	if err := row.Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Source{}, fmt.Errorf("source not found: %w", err)
		}
		return model.Source{}, fmt.Errorf("scan source by name: %w", err)
	}
	return s, nil
}

// ListAll returns all sources ordered by name (or created_at).
func (r *SourceRepository) ListAll(ctx context.Context) ([]model.Source, error) {
	const q = `
SELECT id, name, created_at
FROM sources
ORDER BY name ASC
`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query sources: %w", err)
	}
	defer rows.Close()

	out := make([]model.Source, 0, 8)
	for rows.Next() {
		var s model.Source
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan source row: %w", err)
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}
	return out, nil
}
