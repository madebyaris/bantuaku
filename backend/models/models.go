package models

import (
	"time"
)

// User represents a registered user
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Store represents an UMKM store
type Store struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	StoreName        string    `json:"store_name"`
	Industry         string    `json:"industry,omitempty"`
	Location         string    `json:"location,omitempty"`
	SubscriptionPlan string    `json:"subscription_plan"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// Product represents a product catalog item
type Product struct {
	ID          string    `json:"id"`
	StoreID     string    `json:"store_id"`
	ProductName string    `json:"product_name"`
	SKU         string    `json:"sku,omitempty"`
	Category    string    `json:"category,omitempty"`
	UnitPrice   float64   `json:"unit_price"`
	Cost        float64   `json:"cost,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Sale represents a sales history entry
type Sale struct {
	ID        int64     `json:"id"`
	StoreID   string    `json:"store_id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	SaleDate  time.Time `json:"sale_date"`
	Source    string    `json:"source"` // manual, csv, woocommerce, shopee
	CreatedAt time.Time `json:"created_at"`
}

// Forecast represents a demand forecast for a product
type Forecast struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Forecast30d int       `json:"forecast_30d"`
	Forecast60d int       `json:"forecast_60d,omitempty"`
	Forecast90d int       `json:"forecast_90d,omitempty"`
	Confidence  float64   `json:"confidence"`
	EOQ         float64   `json:"eoq,omitempty"`
	SafetyStock int       `json:"safety_stock,omitempty"`
	Algorithm   string    `json:"algorithm"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Recommendation represents a demand forecast recommendation
type Recommendation struct {
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	ProjectedDemand int    `json:"projected_demand"` // 30-day projected demand
	Reason          string `json:"reason"`
	RiskLevel       string `json:"risk_level"` // low, medium, high
}

// Integration represents an external platform integration
type Integration struct {
	ID           string     `json:"id"`
	StoreID      string     `json:"store_id"`
	Platform     string     `json:"platform"` // woocommerce, shopee, tokopedia
	Status       string     `json:"status"`   // connected, disconnected, error
	LastSync     *time.Time `json:"last_sync,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Metadata     string     `json:"metadata,omitempty"` // JSON string for platform-specific data
	CreatedAt    time.Time  `json:"created_at"`
}

// SentimentData represents sentiment analysis data
type SentimentData struct {
	ProductID      string    `json:"product_id"`
	SentimentScore float64   `json:"sentiment_score"` // -1 to +1
	PositiveCount  int       `json:"positive_count"`
	NegativeCount  int       `json:"negative_count"`
	NeutralCount   int       `json:"neutral_count"`
	RecentMentions []Mention `json:"recent_mentions,omitempty"`
}

// Mention represents a social media mention
type Mention struct {
	Source    string    `json:"source"` // instagram, tiktok, review
	Text      string    `json:"text"`
	Sentiment float64   `json:"sentiment"`
	Date      time.Time `json:"date"`
}

// DashboardSummary represents the main dashboard KPIs
type DashboardSummary struct {
	// Company Info
	CompanyName     string `json:"company_name,omitempty"`
	CompanyIndustry string `json:"company_industry,omitempty"`
	CompanyLocation string `json:"company_location,omitempty"`

	// Revenue Metrics
	RevenueThisMonth  float64 `json:"revenue_this_month"`
	RevenueTrend      float64 `json:"revenue_trend"` // percentage change from last month
	TopSellingProduct string  `json:"top_selling_product,omitempty"`

	// Activity Metrics
	TotalConversations int `json:"total_conversations"`
	TotalInsights      int `json:"total_insights"`
	TotalFileUploads   int `json:"total_file_uploads"`

	// Insights Summary
	InsightsSummary InsightsCounts `json:"insights_summary"`

	// Recent Activity
	RecentConversations []ConversationSummary `json:"recent_conversations,omitempty"`
	RecentFileUploads   []FileUploadSummary   `json:"recent_file_uploads,omitempty"`
}

// InsightsCounts represents counts of each insight type
type InsightsCounts struct {
	Forecast   int `json:"forecast"`
	Market     int `json:"market"`
	Marketing  int `json:"marketing"`
	Regulation int `json:"regulation"`
}

// ConversationSummary represents a brief conversation summary
type ConversationSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	LastMessage string `json:"last_message,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

// FileUploadSummary represents a brief file upload summary
type FileUploadSummary struct {
	ID               string `json:"id"`
	OriginalFilename string `json:"original_filename"`
	SourceType       string `json:"source_type"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

// AIAnalyzeRequest represents a question to the AI assistant
type AIAnalyzeRequest struct {
	Question string `json:"question"`
}

// AIAnalyzeResponse represents the AI assistant's response
type AIAnalyzeResponse struct {
	Answer      string   `json:"answer"`
	Confidence  float64  `json:"confidence"`
	DataSources []string `json:"data_sources,omitempty"`
}
