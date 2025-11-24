package model

import "time"

type Source struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateSourceRequest struct {
	Name string `json:"name" binding:"required"`
}
