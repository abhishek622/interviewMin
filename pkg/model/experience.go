package model

import "time"

type Source string

const (
	SourceLeetcode Source = "leetcode"
	SourceReddit   Source = "reddit"
	SourceGFG      Source = "gfg"
	SourcePersonal Source = "personal"
	SourceOther    Source = "other"
)

type InputType string

const (
	InputTypeURL  InputType = "url"
	InputTypeText InputType = "text"
)

type ProcessStatus string

const (
	ProcessStatusQueued     ProcessStatus = "queued"
	ProcessStatusProcessing ProcessStatus = "processing"
	ProcessStatusCompleted  ProcessStatus = "completed"
	ProcessStatusFailed     ProcessStatus = "failed"
)

type Experience struct {
	ExpID            int64                  `json:"exp_id" db:"exp_id"`
	UserID           string                 `json:"user_id" db:"user_id"`
	InputType        InputType              `json:"input_type" db:"input_type"`
	RawInput         string                 `json:"raw_input" db:"raw_input"`
	InputHash        string                 `json:"input_hash" db:"input_hash"`
	ProcessStatus    ProcessStatus          `json:"process_status" db:"process_status"`
	ProcessError     string                 `json:"process_error" db:"process_error"`
	Attempts         int                    `json:"attempts" db:"attempts"`
	ExtractedTitle   string                 `json:"extracted_title" db:"extracted_title"`
	ExtractedContent string                 `json:"extracted_content" db:"extracted_content"`
	Company          string                 `json:"company" db:"company"`
	Position         string                 `json:"position" db:"position"`
	NoOfRound        int                    `json:"no_of_round" db:"no_of_round"`
	Location         string                 `json:"location" db:"location"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

type CreateExperienceReq struct {
	RawInput  string    `json:"raw_input" binding:"required"`
	InputType InputType `json:"input_type" binding:"required"`
}

type ListExperiencesQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}
