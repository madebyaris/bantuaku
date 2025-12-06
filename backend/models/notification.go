package models

import "time"

// Notification represents a user/company notification
type Notification struct {
	ID        string     `json:"id"`
	CompanyID string     `json:"company_id"`
	UserID    string     `json:"user_id,omitempty"`
	Title     string     `json:"title"`
	Body      string     `json:"body,omitempty"`
	Type      string     `json:"type,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}
