package models

import "time"

type Link struct {
	Slug        string    `json:"slug"`
	Destination string    `json:"destination"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}
