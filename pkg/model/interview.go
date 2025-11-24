package model

import "time"

type Interview struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	SourceID  string    `json:"source_id" db:"source_id"`
	Title     string    `json:"title" db:"title"`
	RawText   string    `json:"raw_text" db:"raw_text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateInterviewRequest struct {
	Title    string `json:"title" binding:"required"`
	RawText  string `json:"raw_text" binding:"required"`
	SourceID string `json:"source_id" binding:"required,uuid"`
}
