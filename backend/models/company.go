package models

import (
	"time"
)

// Company represents a business/UMKM (replaces Store)
type Company struct {
	ID                 string            `json:"id"`
	OwnerUserID        string            `json:"owner_user_id"`
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	Industry           string            `json:"industry,omitempty"`
	BusinessModel      string            `json:"business_model,omitempty"`
	FoundedYear        *int              `json:"founded_year,omitempty"`
	LocationRegion     string            `json:"location_region,omitempty"`
	City               string            `json:"city,omitempty"`
	Country            string            `json:"country"` // Default: "ID"
	Website            string            `json:"website,omitempty"`
	SocialMediaHandles map[string]string `json:"social_media_handles,omitempty"` // JSONB
	Marketplaces       map[string]string `json:"marketplaces,omitempty"`         // JSONB
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// CompanyProfile is an aggregate that merges company data from all sources
type CompanyProfile struct {
	Company     *Company      `json:"company"`
	Products    []*Product    `json:"products"`
	DataSources []*DataSource `json:"data_sources"`
	SalesData   []*Sale       `json:"sales_data,omitempty"` // Aggregated
	LastUpdated time.Time     `json:"last_updated"`
}
