package models

import (
	"time"
)

// FileUpload represents an uploaded file (CSV/XLSX/PDF)
type FileUpload struct {
	ID               string     `json:"id"`
	CompanyID        string     `json:"company_id"`
	UserID           string     `json:"user_id"`
	SourceType       string     `json:"source_type"` // "csv", "xlsx", "pdf"
	OriginalFilename string     `json:"original_filename"`
	StoragePath      string     `json:"storage_path"`
	MimeType         string     `json:"mime_type,omitempty"`
	SizeBytes        int64      `json:"size_bytes"`
	Status           string     `json:"status"` // "uploaded", "processing", "processed", "failed"
	ErrorMessage     string     `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
}

// ExtractedData represents data extracted from a file
type ExtractedData struct {
	Products []ExtractedProduct `json:"products,omitempty"`
	Sales    []ExtractedSale    `json:"sales,omitempty"`
}

// ExtractedProduct represents a product extracted from a file
type ExtractedProduct struct {
	Name     string  `json:"name"`
	Category string  `json:"category,omitempty"`
	Price    float64 `json:"price,omitempty"`
	SKU      string  `json:"sku,omitempty"`
}

// ExtractedSale represents a sale extracted from a file
type ExtractedSale struct {
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Date        time.Time `json:"date"`
}
