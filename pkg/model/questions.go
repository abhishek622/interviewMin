package model

import "time"

type Question struct {
	QID         int64     `json:"q_id" db:"q_id"`
	InterviewID int64     `json:"interview_id" db:"interview_id"`
	Question    string    `json:"question" db:"question"`
	Type        string    `json:"type" db:"type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type QuestionRes struct {
	QID         int64  `json:"q_id"`
	InterviewID int64  `json:"interview_id"`
	Question    string `json:"question"`
	Type        string `json:"type"`
}

type ListQuestionsQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

type CreateQuestionReq struct {
	InterviewID int64  `json:"interview_id"`
	Question    string `json:"question"`
	Type        string `json:"type"`
}

type UpdateQuestionReq struct {
	Question string `json:"question"`
	Type     string `json:"type"`
}
