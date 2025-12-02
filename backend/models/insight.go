package models

import (
	"time"
)

// Insight represents a generated insight (forecast, market, marketing, regulation)
type Insight struct {
	ID           string                 `json:"id"`
	CompanyID    string                 `json:"company_id"`
	Type         string                 `json:"type"`                    // "forecast", "market_prediction", "marketing_recommendation", "gov_regulation"
	InputContext map[string]interface{} `json:"input_context,omitempty"` // JSONB - time ranges, assumptions, filters
	Result       map[string]interface{} `json:"result"`                  // JSONB - numbers, charts, recommended actions
	CreatedAt    time.Time              `json:"created_at"`
}

// ForecastInsightResult represents the result structure for forecast insights
type ForecastInsightResult struct {
	Forecasts []ProductForecast `json:"forecasts"`
}

// ProductForecast represents a forecast for a single product
type ProductForecast struct {
	ProductID   string   `json:"product_id"`
	ProductName string   `json:"product_name"`
	Forecast30D *int     `json:"forecast_30d,omitempty"`
	Forecast60D *int     `json:"forecast_60d,omitempty"`
	Forecast90D *int     `json:"forecast_90d,omitempty"`
	Confidence  float64  `json:"confidence"`
	EOQ         *float64 `json:"eoq,omitempty"`
	SafetyStock *int     `json:"safety_stock,omitempty"`
}

// MarketInsightResult represents the result structure for market prediction insights
type MarketInsightResult struct {
	Scope  string        `json:"scope"` // "local" or "global"
	Trends []MarketTrend `json:"trends"`
}

// MarketTrend represents a market trend
type MarketTrend struct {
	Category   string   `json:"category"`
	TrendName  string   `json:"trend_name"`
	TrendScore float64  `json:"trend_score"`
	GrowthRate *float64 `json:"growth_rate,omitempty"`
	Source     string   `json:"source"` // "marketplace", "google_trends", "social"
}

// MarketingInsightResult represents the result structure for marketing recommendation insights
type MarketingInsightResult struct {
	Recommendations []MarketingRecommendation `json:"recommendations"`
}

// MarketingRecommendation represents a marketing recommendation
type MarketingRecommendation struct {
	CampaignIdea      string             `json:"campaign_idea"`
	Channels          []string           `json:"channels"`
	ContentThemes     []string           `json:"content_themes"`
	TimingSuggestions []string           `json:"timing_suggestions"`
	BudgetAllocation  []BudgetAllocation `json:"budget_allocation,omitempty"`
}

// BudgetAllocation represents budget allocation by channel
type BudgetAllocation struct {
	Channel    string  `json:"channel"`
	Percentage float64 `json:"percentage"`
}

// RegulationInsightResult represents the result structure for government regulation insights
type RegulationInsightResult struct {
	Regulations []Regulation `json:"regulations"`
}

// Regulation represents a government regulation
type Regulation struct {
	Title               string   `json:"title"`
	Category            string   `json:"category"` // "tax", "import", "employment", "food_safety", "other"
	Summary             string   `json:"summary"`
	SourceURL           string   `json:"source_url"`
	EffectiveDate       *string  `json:"effective_date,omitempty"`
	ComplianceChecklist []string `json:"compliance_checklist,omitempty"`
}
