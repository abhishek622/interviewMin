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
	const q = `INSERT INTO questions (interview_id, question, "type") VALUES ($1, $2, $3)`

	for _, question := range questions {
		batch.Queue(q, question.InterviewID, question.Question, question.Type)
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

func (r *Repository) ListQuestionByInterviewID(ctx context.Context, interviewID int64) ([]model.Question, error) {
	const q = `
SELECT q_id, interview_id, question, type, created_at
FROM questions
WHERE interview_id = $1
ORDER BY created_at ASC
`
	rows, err := r.db.Query(ctx, q, interviewID)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var out []model.Question
	for rows.Next() {
		var qs model.Question
		if err := rows.Scan(&qs.QID, &qs.InterviewID, &qs.Question, &qs.Type, &qs.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		out = append(out, qs)
	}
	return out, nil
}

func (r *Repository) UpdateQuestion(ctx context.Context, qID int64, question string, questionType string) error {
	const q = `UPDATE questions SET question = $1, "type" = $2 WHERE q_id = $3`
	_, err := r.db.Exec(ctx, q, question, questionType, qID)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	return nil
}

func (r *Repository) DeleteQuestion(ctx context.Context, qID int64) error {
	const q = `DELETE FROM questions WHERE q_id = $1`
	_, err := r.db.Exec(ctx, q, qID)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	return nil
}

func (r *Repository) CreateQuestion(ctx context.Context, question *model.Question) (*model.Question, error) {
	const q = `INSERT INTO questions (interview_id, question, "type") VALUES ($1, $2, $3) RETURNING q_id`
	err := r.db.QueryRow(ctx, q, question.InterviewID, question.Question, question.Type).Scan(&question.QID)
	if err != nil {
		return nil, fmt.Errorf("insert question: %w", err)
	}
	return question, nil
}
