package model

import "time"

type Question struct {
	QID       int64     `json:"q_id" db:"q_id"`
	ExpID     int64     `json:"exp_id" db:"exp_id"`
	Question  string    `json:"question" db:"question"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type QuestionRes struct {
	QID       int64     `json:"q_id"`
	ExpID     int64     `json:"exp_id"`
	Question  string    `json:"question"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ListQuestionsQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}
