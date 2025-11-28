package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExperienceRepository struct {
	db *pgxpool.Pool
}

func (r *ExperienceRepository) Create(ctx context.Context, exp *model.Experience) error {
	const q = `
INSERT INTO experiences (
	 user_id, company, position, source, no_of_round, 
	source_link, location, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	_, err := r.db.Exec(ctx, q,
		exp.UserID, exp.Company, exp.Position, exp.Source, exp.NoOfRound,
		exp.SourceLink, exp.Location, exp.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert experience: %w", err)
	}
	return nil
}

func (r *ExperienceRepository) GetByID(ctx context.Context, id int64) (*model.Experience, error) {
	const q = `
SELECT 
	exp_id, user_id, company, position, source, no_of_round, 
	source_link, location, metadata, created_at
FROM experiences
WHERE exp_id = $1
`
	var e model.Experience
	row := r.db.QueryRow(ctx, q, id)
	err := row.Scan(
		&e.ExpID, &e.UserID, &e.Company, &e.Position, &e.Source, &e.NoOfRound,
		&e.SourceLink, &e.Location, &e.Metadata, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("experience not found: %w", err)
		}
		return nil, fmt.Errorf("scan experience: %w", err)
	}
	return &e, nil
}

func (r *ExperienceRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]model.Experience, int, error) {
	var total int
	const countQ = `SELECT COUNT(*) FROM experiences WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count experiences: %w", err)
	}

	const q = `
SELECT 
	exp_id, user_id, company, position, source, no_of_round, 
	source_link, location, metadata, created_at
FROM experiences
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query experiences: %w", err)
	}
	defer rows.Close()

	out := make([]model.Experience, 0, limit)
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(
			&e.ExpID, &e.UserID, &e.Company, &e.Position, &e.Source, &e.NoOfRound,
			&e.SourceLink, &e.Location, &e.Metadata, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan experience row: %w", err)
		}
		out = append(out, e)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("rows error: %w", rows.Err())
	}
	return out, total, nil
}
