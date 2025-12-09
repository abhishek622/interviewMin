package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/abhishek622/interviewMin/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateInterview(ctx context.Context, exp *model.Interview) (*int64, error) {
	const q = `
INSERT INTO interviews (
	 user_id, source, raw_input, process_status, metadata
) VALUES ($1, $2, $3, $4, $5) RETURNING interview_id
`
	row := r.db.QueryRow(ctx, q,
		exp.UserID, exp.Source, exp.RawInput, exp.ProcessStatus, exp.Metadata,
	)
	var interviewID int64
	if err := row.Scan(&interviewID); err != nil {
		return nil, fmt.Errorf("insert interview: %w", err)
	}
	return &interviewID, nil
}

func (r *Repository) CreateFullInterview(ctx context.Context, exp *model.Interview) (*int64, error) {
	const q = `
INSERT INTO interviews (
	 user_id, source, raw_input, process_status, company, position, no_of_round, location, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING interview_id
`
	row := r.db.QueryRow(ctx, q,
		exp.UserID, exp.Source, exp.RawInput, exp.ProcessStatus, exp.Company, exp.Position, exp.NoOfRound, exp.Location, exp.Metadata,
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
	interview_id, user_id, source, raw_input, process_status, attempts,
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
		return nil, err
	}
	return &e, nil
}

func (r *Repository) ListInterviewByUser(ctx context.Context, userID uuid.UUID, limit, offset int, filters map[string]interface{}, search *string) ([]model.Interview, int, error) {
	// Base Query Construction
	whereConditions := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIndex := 2
	fmt.Println(filters)
	if len(filters) > 0 {
		for col, val := range filters {
			fmt.Println(col, val)
			whereConditions = append(whereConditions, fmt.Sprintf("%s = ANY($%d)", col, argIndex))
			args = append(args, val)
			argIndex++
		}
	}

	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		whereConditions = append(whereConditions, fmt.Sprintf("(company ILIKE $%d OR position ILIKE $%d)", argIndex, argIndex+1))
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	whereClause := strings.Join(whereConditions, " AND ")
	fmt.Println(whereClause)
	// 1. Get Total Count
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(1) FROM interviews WHERE %s", whereClause)
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count interview: %w", err)
	}

	// 2. Get Data
	listQ := fmt.Sprintf(`SELECT 
	interview_id, user_id, source, raw_input, process_status, attempts, 
	process_error, company, position, no_of_round, location, metadata,
	created_at, updated_at FROM interviews WHERE %s
	ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	listArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query interview: %w", err)
	}
	defer rows.Close()

	out := make([]model.Interview, 0, limit)
	for rows.Next() {
		var e model.Interview
		if err := rows.Scan(
			&e.InterviewID, &e.UserID, &e.Source, &e.RawInput, &e.ProcessStatus, &e.Attempts,
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

func (r *Repository) ListInterviewByUserStats(ctx context.Context, userID uuid.UUID) (*model.InterviewListStats, error) {

	query := `
		WITH base_data AS (
			SELECT source, process_status
			FROM interviews
			WHERE user_id = $1
		),
		source_counts AS (
			SELECT source, COUNT(*) AS count
			FROM base_data
			GROUP BY source
		),
		status_counts AS (
			SELECT process_status, COUNT(*) AS count
			FROM base_data
			GROUP BY process_status
		)
		SELECT 
			(SELECT json_agg(json_build_object('field', source, 'count', count) ORDER BY count DESC) FROM source_counts) AS source_stats,
			(SELECT json_agg(json_build_object('field', process_status, 'count', count) ORDER BY count DESC) FROM status_counts) AS status_stats
	`

	var sourceStatsJSON, statusStatsJSON []byte

	err := r.db.QueryRow(ctx, query, userID).Scan(&sourceStatsJSON, &statusStatsJSON)
	if err != nil {
		return nil, fmt.Errorf("query stats: %w", err)
	}

	var sourceStats []model.ListStats
	if sourceStatsJSON != nil {
		if err := json.Unmarshal(sourceStatsJSON, &sourceStats); err != nil {
			return nil, fmt.Errorf("unmarshal source stats: %w", err)
		}
	}

	var statusStats []model.ListStats
	if statusStatsJSON != nil {
		if err := json.Unmarshal(statusStatsJSON, &statusStats); err != nil {
			return nil, fmt.Errorf("unmarshal status stats: %w", err)
		}
	}

	// Helper to fill missing keys with 0
	fillMissing := func(current []model.ListStats, allKeys []string) []model.ListStats {
		m := make(map[string]int)
		for _, item := range current {
			m[item.Field] = item.Count
		}

		out := make([]model.ListStats, 0, len(allKeys))
		for _, key := range allKeys {
			out = append(out, model.ListStats{
				Field: key,
				Count: m[key],
			})
		}
		return out
	}

	allSources := []string{
		string(model.SourceLeetcode),
		string(model.SourceReddit),
		string(model.SourceGFG),
		string(model.SourceOther),
		string(model.SourcePersonal),
	}

	allStatuses := []string{
		string(model.ProcessStatusQueued),
		string(model.ProcessStatusProcessing),
		string(model.ProcessStatusCompleted),
		string(model.ProcessStatusFailed),
	}

	result := &model.InterviewListStats{
		SourceStats:        fillMissing(sourceStats, allSources),
		ProcessStatusStats: fillMissing(statusStats, allStatuses),
	}

	return result, nil
}

func (r *Repository) DeleteInterview(ctx context.Context, interviewID int64) error {
	return r.execTx(ctx, func(tx pgx.Tx) error {
		const q2 = `DELETE FROM questions WHERE interview_id = $1`
		_, err := tx.Exec(ctx, q2, interviewID)
		if err != nil {
			return fmt.Errorf("delete questions: %w", err)
		}

		const q = `DELETE FROM interviews WHERE interview_id = $1`
		_, err = tx.Exec(ctx, q, interviewID)
		if err != nil {
			return fmt.Errorf("delete interview: %w", err)
		}

		return nil
	})
}

func (r *Repository) DeleteInterviews(ctx context.Context, interviewIDs []int64) error {
	return r.execTx(ctx, func(tx pgx.Tx) error {
		const qQuestions = `DELETE FROM questions WHERE interview_id = ANY($1)`
		_, err := tx.Exec(ctx, qQuestions, interviewIDs)
		if err != nil {
			return fmt.Errorf("delete questions: %w", err)
		}

		const qInterviews = `DELETE FROM interviews WHERE interview_id = ANY($1)`
		_, err = tx.Exec(ctx, qInterviews, interviewIDs)
		if err != nil {
			return fmt.Errorf("delete interview: %w", err)
		}

		return nil
	})
}

func (r *Repository) CheckInterviewExists(ctx context.Context, interviewIDs []int64) (int, error) {
	var count int
	const q = `SELECT COUNT(interview_id) FROM interviews WHERE interview_id = ANY($1)`
	if err := r.db.QueryRow(ctx, q, interviewIDs).Scan(&count); err != nil {
		return 0, fmt.Errorf("check interview exists: %w", err)
	}
	return count, nil
}

func (r *Repository) GetInterviewStats(ctx context.Context, userID uuid.UUID) (model.InterviewStats, error) {
	const q = `
    SELECT
        -- 1. Counts 
        COUNT(*) AS total_6_months,
        COUNT(*) FILTER (WHERE source = 'personal') AS personal_6_months,

        -- 2. This Month vs Last Month 
        COUNT(*) FILTER (WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE)) AS total_this_month,
        COUNT(*) FILTER (WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') 
                           AND created_at < DATE_TRUNC('month', CURRENT_DATE)) AS total_last_month,

        COUNT(*) FILTER (WHERE source = 'personal' AND created_at >= DATE_TRUNC('month', CURRENT_DATE)) AS personal_this_month,
        COUNT(*) FILTER (WHERE source = 'personal' AND created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') 
                           AND created_at < DATE_TRUNC('month', CURRENT_DATE)) AS personal_last_month,

        -- 3. Top Companies
        COALESCE((
            SELECT array_agg(t.company)
            FROM (
                SELECT company
                FROM interviews
                WHERE user_id = $1 
                  AND company IS NOT NULL 
                  AND company != ''
                  AND created_at >= CURRENT_DATE - INTERVAL '6 months'
                GROUP BY company
                ORDER BY COUNT(*) DESC
                LIMIT 5
            ) t
        ), '{}') AS top_companies
    FROM interviews
    WHERE user_id = $1 
      AND created_at >= CURRENT_DATE - INTERVAL '6 months';
    `

	var (
		stats             model.InterviewStats
		totalThisMonth    int
		totalLastMonth    int
		personalThisMonth int
		personalLastMonth int
	)

	err := r.db.QueryRow(ctx, q, userID).Scan(
		&stats.Total,
		&stats.Personal,
		&totalThisMonth,
		&totalLastMonth,
		&personalThisMonth,
		&personalLastMonth,
		&stats.TopCompanies,
	)

	if err != nil {
		return stats, fmt.Errorf("get interview stats: %w", err)
	}

	if stats.TopCompanies == nil {
		stats.TopCompanies = []string{}
	}

	// Calculate Percentage Changes
	stats.TotalChange = calculateGrowth(totalThisMonth, totalLastMonth)
	stats.PersonalChange = calculateGrowth(personalThisMonth, personalLastMonth)

	return stats, nil
}

// Helper function to calculate percentage growth safely
func calculateGrowth(current, previous int) int {
	if previous == 0 {
		if current > 0 {
			return 100 // 0 to 5 = 100% growth (technically infinite, but 100 is standard for UI)
		}
		return 0 // 0 to 0 = 0% change
	}
	// Formula: ((Current - Previous) / Previous) * 100
	// We convert to float for division, then round back to int
	delta := float64(current - previous)
	return int((delta / float64(previous)) * 100)
}
