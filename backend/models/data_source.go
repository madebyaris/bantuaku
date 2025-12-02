package models

import (
	"time"
)

// DataSource represents an external data connection/channel
type DataSource struct {
	ID        string                 `json:"id"`
	CompanyID string                 `json:"company_id"`
	Type      string                 `json:"type"`           // "manual", "csv", "xlsx", "pdf", "marketplace", "google_trends", "regulation"
	Provider  string                 `json:"provider"`       // "tokopedia", "shopee", "bukalapak", "google_trends", "peraturan_go_id"
	Meta      map[string]interface{} `json:"meta,omitempty"` // JSONB - account name, URLs, etc.
	Status    string                 `json:"status"`         // "active", "inactive", "error"
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}
