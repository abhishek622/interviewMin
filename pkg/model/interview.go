package model

import (
	"time"

	"github.com/google/uuid"
)

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
	UserID        uuid.UUID              `json:"user_id" db:"user_id"`
	Source        Source                 `json:"source" db:"source"`
	RawInput      string                 `json:"raw_input" db:"raw_input"`
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

type CreateInterviewWithAIReq struct {
	RawInput string `json:"raw_input" binding:"required"`
	Source   Source `json:"source" binding:"required"`
}

type CreateInterviewReq struct {
	Source    Source  `json:"source" binding:"required"`
	Company   string  `json:"company" binding:"required"`
	Position  string  `json:"position" binding:"required"`
	NoOfRound *int    `json:"no_of_round"`
	Location  *string `json:"location"`
	RawInput  string  `json:"raw_input" binding:"required"`
}

type Filter struct {
	Source        *[]Source        `form:"source"`
	ProcessStatus *[]ProcessStatus `form:"process_status"`
}

type ListInterviewQuery struct {
	Page     int     `form:"page,default=1"`
	PageSize int     `form:"page_size,default=20"`
	Search   *string `form:"search"`
	Filter   *Filter `form:"filter"`
}

type DeleteInterviewsRequest struct {
	InterviewIDs []int64 `json:"interview_ids" binding:"required,min=1,max=100"`
}

type PatchInterviewRequest struct {
	Company        *string `json:"company,omitempty"`
	Position       *string `json:"position,omitempty"`
	NoOfRound      *int    `json:"no_of_round,omitempty" binding:"min=1,max=100"`
	Location       *string `json:"location,omitempty"`
	Title          *string `json:"title,omitempty"`
	FullExperience *string `json:"full_experience,omitempty"`
}

type InterviewStats struct {
	Total          int      `json:"total"`
	TotalChange    int      `json:"total_change"` // Implied % change
	Personal       int      `json:"personal"`
	PersonalChange int      `json:"personal_change"` // Implied % change
	TopCompanies   []string `json:"top_companies"`
}

type ListStats struct {
	Field string `json:"field"`
	Count int    `json:"count"`
}

type InterviewListStats struct {
	SourceStats        []ListStats `json:"source"`
	ProcessStatusStats []ListStats `json:"process_status"`
}
