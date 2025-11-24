package model

import "time"

type Entry struct {
	ID          string                 `json:"id" db:"id"`
	InterviewID string                 `json:"interview_id" db:"interview_id"`
	Title       string                 `json:"title" db:"title"`
	Content     map[string]interface{} `json:"content" db:"content"` // JSONB
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

type EntryResponse struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   map[string]interface{} `json:"content"`
	CreatedAt time.Time              `json:"created_at"`
}

type ConvertInterviewRequest struct {
	OverrideTitle string `json:"override_title"`
}

type CreateEntryResponse struct {
	ID string `json:"id"`
}

type ListInterviewsQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

type ListEntriesQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}
