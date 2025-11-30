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

type Experience struct {
	ExpID      int64                  `json:"exp_id" db:"exp_id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Company    string                 `json:"company" db:"company"`
	Position   string                 `json:"position" db:"position"`
	Source     Source                 `json:"source" db:"source"`
	NoOfRound  int                    `json:"no_of_round" db:"no_of_round"`
	SourceLink string                 `json:"source_link" db:"source_link"`
	Location   string                 `json:"location" db:"location"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

type CreateExperienceRequest struct {
	SourceLink string `json:"source_link" binding:"optional"`
	Source     Source `json:"source" binding:"required"`
	TextInput  string `json:"text_input" binding:"optional"`
}

type ListExperiencesQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}
