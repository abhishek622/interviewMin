package model

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	CompanyID uuid.UUID `json:"company_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CompanyList struct {
	CompanyID       uuid.UUID `json:"company_id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	TotalInterviews int       `json:"total_interviews"`
}

type CompanyListReq struct {
	Limit  int    `json:"limit" form:"limit,default=20"`
	Offset int    `json:"offset" form:"offset,default=0"`
	Sort   string `json:"sort" form:"sort,default=created_at"`
}

type CompanyDetails struct {
	CompanyID       uuid.UUID `json:"company_id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	TotalInterviews int       `json:"total_interviews"`
	AvgRounds       float64   `json:"avg_rounds"`
}

type CompanyListNameList struct {
	CompanyID uuid.UUID `json:"company_id"`
	Name      string    `json:"name"`
}
