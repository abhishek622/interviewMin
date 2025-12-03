package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateInterview(ctx context.Context, exp *model.Interview) (*int64, error) {
	const q = `
INSERT INTO interviews (
	 user_id, source, raw_input, input_hash, process_status, metadata
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING interview_id
`
	row := r.db.QueryRow(ctx, q,
		exp.UserID, exp.Source, exp.RawInput, exp.InputHash, exp.ProcessStatus, exp.Metadata,
	)
	var interviewID int64
	if err := row.Scan(&interviewID); err != nil {
		return nil, fmt.Errorf("insert interview: %w", err)
	}
	return &interviewID, nil
}

func (r *Repository) UpdateInterview(ctx context.Context, interviewID int64, updates map[string]interface{}) error {
	validCols := map[string]bool{
		"process_status": true, "process_error": true, "company": true,
		"position": true, "source": true, "no_of_round": true,
		"location": true, "metadata": true,
	}

	query := "UPDATE interviews SET "
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

	query += fmt.Sprintf(" WHERE interview_id = $%d", argId)
	args = append(args, interviewID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update interview: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("interview not found")
	}

	return nil
}

func (r *Repository) GetInterviewByID(ctx context.Context, interviewID int64) (*model.Interview, error) {
	const q = `
SELECT 
	interview_id, source, raw_input, process_status, attempts,
	process_error, company, position, no_of_round, location, metadata,
	created_at, updated_at FROM interviews WHERE interview_id = $1
`
	var e model.Interview
	row := r.db.QueryRow(ctx, q, interviewID)
	err := row.Scan(
		&e.InterviewID, &e.UserID, &e.Source, &e.RawInput, &e.ProcessStatus, &e.Attempts,
		&e.ProcessError, &e.Company, &e.Position, &e.NoOfRound, &e.Location, &e.Metadata, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("interview not found: %w", err)
		}
		return nil, fmt.Errorf("scan interview: %w", err)
	}
	return &e, nil
}

func (r *Repository) ListInterviewByUser(ctx context.Context, userID string, limit, offset int) ([]model.Interview, int, error) {
	var total int
	const countQ = `SELECT COUNT(*) FROM interviews WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count interview: %w", err)
	}

	const q = `
SELECT 
	interview_id, source, raw_input, process_status, attempts, 
	process_error, company, position, no_of_round, location, metadata,
	created_at, updated_at FROM interviews WHERE user_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3
`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query interview: %w", err)
	}
	defer rows.Close()

	out := make([]model.Interview, 0, limit)
	for rows.Next() {
		var e model.Interview
		if err := rows.Scan(
			&e.InterviewID, &e.Source, &e.RawInput, &e.ProcessStatus, &e.Attempts,
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
