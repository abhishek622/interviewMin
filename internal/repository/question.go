package repository

import (
	"context"
	"fmt"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateQuestions(ctx context.Context, questions []model.Question) error {
	if len(questions) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const q = `INSERT INTO questions (exp_id, question, "type") VALUES ($1, $2, $3)`

	for _, question := range questions {
		batch.Queue(q, question.ExpID, question.Question, question.Type)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	// Execute each queued statement
	for i := 0; i < len(questions); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch insert question %d: %w", i, err)
		}
	}

	return nil
}

func (r *Repository) ListQuestionByInterviewID(ctx context.Context, expID int64) ([]model.Question, error) {
	const q = `
SELECT q_id, exp_id, question, type, created_at
FROM questions
WHERE exp_id = $1
ORDER BY created_at ASC
`
	rows, err := r.db.Query(ctx, q, expID)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var out []model.Question
	for rows.Next() {
		var qs model.Question
		if err := rows.Scan(&qs.QID, &qs.ExpID, &qs.Question, &qs.Type, &qs.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		out = append(out, qs)
	}
	return out, nil
}
