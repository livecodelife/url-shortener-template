package models

import "time"

type Click struct {
	Slug      string    `json:"slug"`
	ClickedAt time.Time `json:"clicked_at"`
	UserAgent *string   `json:"user_agent"`
	Referer   *string   `json:"referer"`
	IPAddress *string   `json:"ip_address"`
}
