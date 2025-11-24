package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EntryRepository is the concrete implementation for entries.
type EntryRepository struct {
	db *pgxpool.Pool
}

// Create inserts an entry (structured data) and returns its id
func (r *EntryRepository) Create(ctx context.Context, interviewID, title string, content map[string]interface{}) (string, error) {
	id := uuid.New().String()
	b, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal content: %w", err)
	}
	const q = `
INSERT INTO entries (id, interview_id, title, content, created_at)
VALUES ($1, $2, $3, $4::jsonb, now())
`
	_, err = r.db.Exec(ctx, q, id, interviewID, title, b)
	if err != nil {
		return "", fmt.Errorf("insert entry: %w", err)
	}
	return id, nil
}

// GetByID fetches an entry by id
func (r *EntryRepository) GetByID(ctx context.Context, id string) (model.Entry, error) {
	const q = `
SELECT id, interview_id, title, content, created_at
FROM entries
WHERE id = $1
`
	var e model.Entry
	var contentBytes []byte
	row := r.db.QueryRow(ctx, q, id)
	if err := row.Scan(&e.ID, &e.InterviewID, &e.Title, &contentBytes, &e.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Entry{}, fmt.Errorf("entry not found: %w", err)
		}
		return model.Entry{}, fmt.Errorf("scan entry: %w", err)
	}
	if len(contentBytes) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(contentBytes, &m); err != nil {
			return model.Entry{}, fmt.Errorf("unmarshal content: %w", err)
		}
		e.Content = m
	} else {
		e.Content = map[string]interface{}{}
	}
	return e, nil
}

// ListByUser returns entries for a user's interviews (joins interviews -> entries)
// This ensures entries belong to the user.
func (r *EntryRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.Entry, int, error) {
	// total count
	var total int
	const countQ = `
SELECT COUNT(e.id)
FROM entries e
JOIN interviews i ON e.interview_id = i.id
WHERE i.user_id = $1
`
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count entries: %w", err)
	}

	const q = `
SELECT e.id, e.interview_id, e.title, e.content, e.created_at
FROM entries e
JOIN interviews i ON e.interview_id = i.id
WHERE i.user_id = $1
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3
`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	out := make([]model.Entry, 0, 8)
	for rows.Next() {
		var e model.Entry
		var contentBytes []byte
		if err := rows.Scan(&e.ID, &e.InterviewID, &e.Title, &contentBytes, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan entry row: %w", err)
		}
		if len(contentBytes) > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal(contentBytes, &m); err != nil {
				return nil, 0, fmt.Errorf("unmarshal content: %w", err)
			}
			e.Content = m
		} else {
			e.Content = map[string]interface{}{}
		}
		out = append(out, e)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("rows error: %w", rows.Err())
	}
	return out, total, nil
}

// SaveContent updates the content JSONB for an entry
func (r *EntryRepository) SaveContent(ctx context.Context, entryID string, content map[string]interface{}) error {
	b, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}
	const q = `UPDATE entries SET content = $1::jsonb WHERE id = $2`
	ct, err := r.db.Exec(ctx, q, b, entryID)
	if err != nil {
		return fmt.Errorf("update entry content: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("entry not found")
	}
	return nil
}
