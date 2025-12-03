package model

import "time"

type Source string

const (
	SourceLeetcode Source = "leetcode"
	SourceReddit   Source = "reddit"
	SourceGFG      Source = "geeksforgeeks"
	SourceOther    Source = "other"
	SourcePersonal Source = "personal"
)

type ProcessStatus string

const (
	ProcessStatusQueued     ProcessStatus = "queued"
	ProcessStatusProcessing ProcessStatus = "processing"
	ProcessStatusCompleted  ProcessStatus = "completed"
	ProcessStatusFailed     ProcessStatus = "failed"
)

type Interview struct {
	InterviewID   int64                  `json:"interview_id" db:"interview_id"`
	UserID        string                 `json:"user_id" db:"user_id"`
	Source        Source                 `json:"source" db:"source"`
	RawInput      string                 `json:"raw_input" db:"raw_input"`
	InputHash     string                 `json:"input_hash" db:"input_hash"`
	ProcessStatus ProcessStatus          `json:"process_status" db:"process_status"`
	ProcessError  *string                `json:"process_error" db:"process_error"`
	Attempts      *int                   `json:"attempts" db:"attempts"`
	Company       *string                `json:"company" db:"company"`
	Position      *string                `json:"position" db:"position"`
	NoOfRound     *int                   `json:"no_of_round" db:"no_of_round"`
	Location      *string                `json:"location" db:"location"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

type CreateInterviewReq struct {
	RawInput string `json:"raw_input" binding:"required"`
	Source   Source `json:"source" binding:"required"`
}

type ListInterviewQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}
