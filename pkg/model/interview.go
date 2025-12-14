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
	ProcessStatusSuccess ProcessStatus = "success"
	ProcessStatusFailed  ProcessStatus = "failed"
)

type Interview struct {
	InterviewID   int64                  `json:"interview_id" db:"interview_id"`
	CompanyID     uuid.UUID              `json:"company_id" db:"company_id"`
	UserID        uuid.UUID              `json:"user_id" db:"user_id"`
	Source        Source                 `json:"source" db:"source"`
	RawInput      string                 `json:"raw_input" db:"raw_input"`
	ProcessStatus ProcessStatus          `json:"process_status" db:"process_status"`
	ProcessError  *string                `json:"process_error" db:"process_error"`
	Attempts      *int                   `json:"attempts" db:"attempts"`
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
	Source    Source     `json:"source" binding:"required"`
	CompanyID *uuid.UUID `json:"company_id"`
	Company   string     `json:"company" binding:"required"`
	Position  string     `json:"position" binding:"required"`
	NoOfRound *int       `json:"no_of_round"`
	Location  *string    `json:"location"`
	RawInput  string     `json:"raw_input" binding:"required"`
}

type Filter struct {
	Source        *[]Source        `json:"source" form:"source"`
	ProcessStatus *[]ProcessStatus `json:"process_status" form:"process_status"`
}

type ListInterviewQuery struct {
	Page      int       `json:"page" form:"page,default=1"`
	PageSize  int       `json:"page_size" form:"page_size,default=20"`
	Search    *string   `json:"search" form:"search"`
	CompanyID uuid.UUID `json:"company_id" form:"company_id" binding:"required"`
	Filter    *Filter   `json:"filter" form:"filter"`
}

type DeleteInterviewsRequest struct {
	InterviewIDs []int64 `json:"interview_ids" binding:"required,min=1,max=100"`
}

type PatchInterviewRequest struct {
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	Company   *string    `json:"company,omitempty"`
	Position  *string    `json:"position,omitempty"`
	NoOfRound *int       `json:"no_of_round,omitempty" binding:"min=1,max=100"`
	Location  *string    `json:"location,omitempty"`
	Title     *string    `json:"title,omitempty"`
	RawInput  *string    `json:"raw_input,omitempty"`
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

type RecentInterviews struct {
	InterviewID   int64         `json:"interview_id"`
	Source        Source        `json:"source"`
	ProcessStatus ProcessStatus `json:"process_status"`
	Position      *string       `json:"position"`
	NoOfRound     *int          `json:"no_of_round"`
	Location      *string       `json:"location"`
	CreatedAt     time.Time     `json:"created_at"`
	CompanyID     uuid.UUID     `json:"company_id"`
	CompanyName   string        `json:"company_name"`
}

type InterviewRes struct {
	InterviewID   int64                  `json:"interview_id"`
	CompanyID     uuid.UUID              `json:"company_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Source        Source                 `json:"source"`
	RawInput      string                 `json:"raw_input"`
	ProcessStatus ProcessStatus          `json:"process_status"`
	ProcessError  *string                `json:"process_error"`
	Position      *string                `json:"position"`
	NoOfRound     *int                   `json:"no_of_round"`
	Location      *string                `json:"location"`
	CreatedAt     time.Time              `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	CompanyName   *string                `json:"company_name"`
}

type InterviewListItem struct {
	InterviewID   int64                  `json:"interview_id"`
	CompanyID     uuid.UUID              `json:"company_id"`
	CompanyName   string                 `json:"company_name"`
	Source        Source                 `json:"source"`
	RawInput      string                 `json:"raw_input"`
	ProcessStatus ProcessStatus          `json:"process_status"`
	ProcessError  *string                `json:"process_error"`
	Position      *string                `json:"position"`
	NoOfRound     *int                   `json:"no_of_round"`
	Location      *string                `json:"location"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
}
