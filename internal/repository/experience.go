package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateExperience(ctx context.Context, exp *model.Experience) (*int64, error) {
	const q = `
INSERT INTO experiences (
	 user_id, source, raw_input, input_hash, process_status, metadata
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING exp_id
`
	row := r.db.QueryRow(ctx, q,
		exp.UserID, exp.Source, exp.RawInput, exp.InputHash, exp.ProcessStatus, exp.Metadata,
	)
	var expID int64
	if err := row.Scan(&expID); err != nil {
		return nil, fmt.Errorf("insert experience: %w", err)
	}
	return &expID, nil
}

func (r *Repository) UpdateExperience(ctx context.Context, expID int64, updates map[string]interface{}) error {
	validCols := map[string]bool{
		"process_status": true, "process_error": true, "company": true,
		"position": true, "source": true, "no_of_round": true,
		"location": true, "metadata": true,
	}

	query := "UPDATE experiences SET "
	args := []interface{}{}
	argId := 1

	for col, val := range updates {
		if !validCols[col] {
			continue // Skip invalid columns
		}

		if argId > 1 {
			query += ", "
		}

		// Append "column_name = $1", "column_name = $2", etc.
		query += fmt.Sprintf("%s = $%d", col, argId)
		args = append(args, val)
		argId++
	}

	query += fmt.Sprintf(" WHERE exp_id = $%d", argId)
	args = append(args, expID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update experience: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("experience not found")
	}

	return nil
}

func (r *Repository) GetExperienceByID(ctx context.Context, id int64) (*model.Experience, error) {
	const q = `
SELECT 
	exp_id, user_id, source, raw_input, process_status, attempts, 
	COALESCE(process_error, ''), 
	COALESCE(company, ''), 
	COALESCE(position, ''), 
	COALESCE(no_of_round, 0), 
	COALESCE(location, ''), 
	metadata, created_at, updated_at
FROM experiences WHERE exp_id = $1
`
	var e model.Experience
	row := r.db.QueryRow(ctx, q, id)
	err := row.Scan(
		&e.ExpID, &e.UserID, &e.Source, &e.RawInput, &e.ProcessStatus, &e.Attempts,
		&e.ProcessError, &e.Company, &e.Position, &e.NoOfRound, &e.Location, &e.Metadata, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("experience not found: %w", err)
		}
		return nil, fmt.Errorf("scan experience: %w", err)
	}
	return &e, nil
}

func (r *Repository) ListExperienceByUser(ctx context.Context, userID string, limit, offset int) ([]model.Experience, int, error) {
	var total int
	const countQ = `SELECT COUNT(*) FROM experiences WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count experiences: %w", err)
	}

	const q = `
SELECT 
	exp_id, source, raw_input, process_status, attempts, 
	COALESCE(process_error, ''), 
	COALESCE(company, ''), 
	COALESCE(position, ''), 
	COALESCE(no_of_round, 0), 
	COALESCE(location, ''), 
	metadata,
	created_at,
	updated_at
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
			&e.ExpID, &e.Source, &e.RawInput, &e.ProcessStatus, &e.Attempts,
			&e.ProcessError, &e.Company, &e.Position, &e.NoOfRound, &e.Location, &e.Metadata, &e.CreatedAt, &e.UpdatedAt,
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
