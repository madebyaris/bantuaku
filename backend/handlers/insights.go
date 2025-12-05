package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/models"
	"github.com/bantuaku/backend/validation"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ForecastInsightRequest represents a request for forecast insights
type ForecastInsightRequest struct {
	CompanyID     string   `json:"company_id" validate:"required"`
	HorizonMonths int      `json:"horizon_months" validate:"required,oneof=1 2 3"` // 1=30d, 2=60d, 3=90d
	ProductIDs    []string `json:"product_ids,omitempty"`
	Channels      []string `json:"channels,omitempty"`
}

// MarketInsightRequest represents a request for market prediction insights
type MarketInsightRequest struct {
	CompanyID      string   `json:"company_id" validate:"required"`
	Scope          string   `json:"scope" validate:"required,oneof=local global"`
	TargetProducts []string `json:"target_products,omitempty"`
	Categories     []string `json:"categories,omitempty"`
}

// MarketingInsightRequest represents a request for marketing recommendation insights
type MarketingInsightRequest struct {
	CompanyID      string   `json:"company_id" validate:"required"`
	TargetProducts []string `json:"target_products,omitempty"`
	BudgetRange    *struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"budget_range,omitempty"`
	MainChannels []string `json:"main_channels,omitempty"`
}

// RegulationInsightRequest represents a request for government regulation insights
type RegulationInsightRequest struct {
	CompanyID string `json:"company_id" validate:"required"`
	Industry  string `json:"industry,omitempty"`
	Region    string `json:"region,omitempty"`
	LegalForm string `json:"legal_form,omitempty"`
}

// InsightResponse represents a generic insight response
type InsightResponse struct {
	InsightID string                 `json:"insight_id"`
	Type      string                 `json:"type"`
	Result    map[string]interface{} `json:"result"`
	CreatedAt time.Time              `json:"created_at"`
}

// GenerateForecastInsight generates forecast insights
func (h *Handler) GenerateForecastInsight(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())

	var req ForecastInsightRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Use company_id from context if not provided in request
	if req.CompanyID == "" {
		req.CompanyID = companyID
	}

	// TODO: Implement forecast generation using CompanyProfile
	// For now, return placeholder response
	insightID := uuid.New().String()
	result := map[string]interface{}{
		"forecasts": []models.ProductForecast{},
		"message":   "Forecast akan dihasilkan setelah data penjualan tersedia. Silakan input data melalui AI Assistant.",
	}

	// Store insight in database
	inputContext, _ := json.Marshal(req)
	resultJSON, _ := json.Marshal(result)
	ctx := r.Context()
	h.db.Pool().Exec(ctx, `
		INSERT INTO insights (id, company_id, type, input_context, result, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, insightID, req.CompanyID, "forecast", string(inputContext), string(resultJSON))

	h.respondJSON(w, http.StatusOK, InsightResponse{
		InsightID: insightID,
		Type:      "forecast",
		Result:    result,
		CreatedAt: time.Now(),
	})
}

// GenerateMarketInsight generates market prediction insights
func (h *Handler) GenerateMarketInsight(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())

	var req MarketInsightRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Use company_id from context if not provided in request
	if req.CompanyID == "" {
		req.CompanyID = companyID
	}

	// TODO: Implement market prediction using connectors (marketplace, Google Trends)
	// For now, return placeholder response
	insightID := uuid.New().String()
	result := map[string]interface{}{
		"scope":   req.Scope,
		"trends":  []models.MarketTrend{},
		"message": "Prediksi pasar akan dihasilkan setelah koneksi data eksternal tersedia.",
	}

	// Store insight in database
	inputContext, _ := json.Marshal(req)
	resultJSON, _ := json.Marshal(result)
	ctx := r.Context()
	h.db.Pool().Exec(ctx, `
		INSERT INTO insights (id, company_id, type, input_context, result, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, insightID, req.CompanyID, "market_prediction", string(inputContext), string(resultJSON))

	h.respondJSON(w, http.StatusOK, InsightResponse{
		InsightID: insightID,
		Type:      "market_prediction",
		Result:    result,
		CreatedAt: time.Now(),
	})
}

// GenerateMarketingInsight generates marketing recommendation insights
func (h *Handler) GenerateMarketingInsight(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())

	var req MarketingInsightRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Use company_id from context if not provided in request
	if req.CompanyID == "" {
		req.CompanyID = companyID
	}

	// TODO: Implement marketing recommendation using AI + CompanyProfile + market data
	// For now, return placeholder response
	insightID := uuid.New().String()
	result := map[string]interface{}{
		"recommendations": []models.MarketingRecommendation{},
		"message":         "Rekomendasi marketing akan dihasilkan setelah AI Assistant mengumpulkan informasi tentang bisnis Anda.",
	}

	// Store insight in database
	inputContext, _ := json.Marshal(req)
	resultJSON, _ := json.Marshal(result)
	ctx := r.Context()
	h.db.Pool().Exec(ctx, `
		INSERT INTO insights (id, company_id, type, input_context, result, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, insightID, req.CompanyID, "marketing_recommendation", string(inputContext), string(resultJSON))

	h.respondJSON(w, http.StatusOK, InsightResponse{
		InsightID: insightID,
		Type:      "marketing_recommendation",
		Result:    result,
		CreatedAt: time.Now(),
	})
}

// GenerateRegulationInsight generates government regulation insights
func (h *Handler) GenerateRegulationInsight(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())

	var req RegulationInsightRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	if err := validation.Validate(&req); err != nil {
		h.respondError(w, err, r)
		return
	}

	// Use company_id from context if not provided in request
	if req.CompanyID == "" {
		req.CompanyID = companyID
	}

	// TODO: Implement regulation fetching using connectors (Indonesia regulation scraper)
	// For now, return placeholder response
	insightID := uuid.New().String()
	result := map[string]interface{}{
		"regulations": []models.Regulation{},
		"message":     "Informasi peraturan akan ditampilkan setelah AI Assistant mengetahui industri dan lokasi bisnis Anda.",
	}

	// Store insight in database
	inputContext, _ := json.Marshal(req)
	resultJSON, _ := json.Marshal(result)
	ctx := r.Context()
	h.db.Pool().Exec(ctx, `
		INSERT INTO insights (id, company_id, type, input_context, result, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, insightID, req.CompanyID, "gov_regulation", string(inputContext), string(resultJSON))

	h.respondJSON(w, http.StatusOK, InsightResponse{
		InsightID: insightID,
		Type:      "gov_regulation",
		Result:    result,
		CreatedAt: time.Now(),
	})
}

// GetInsights retrieves insight history for a company
func (h *Handler) GetInsights(w http.ResponseWriter, r *http.Request) {
	companyID := middleware.GetCompanyID(r.Context())
	insightType := r.URL.Query().Get("type")

	// Use company_id from context if not provided in query
	queryCompanyID := r.URL.Query().Get("company_id")
	if queryCompanyID == "" {
		queryCompanyID = companyID
	}

	if queryCompanyID == "" {
		h.respondError(w, errors.NewValidationError("company_id is required", ""), r)
		return
	}

	ctx := r.Context()
	var rows pgx.Rows
	var err error

	if insightType != "" {
		// Filter by type
		rows, err = h.db.Pool().Query(ctx, `
			SELECT id, company_id, type, input_context, result, created_at
			FROM insights
			WHERE company_id = $1 AND type = $2
			ORDER BY created_at DESC
			LIMIT 100
		`, queryCompanyID, insightType)
	} else {
		// Get all insights
		rows, err = h.db.Pool().Query(ctx, `
			SELECT id, company_id, type, input_context, result, created_at
			FROM insights
			WHERE company_id = $1
			ORDER BY created_at DESC
			LIMIT 100
		`, queryCompanyID)
	}

	if err != nil {
		h.respondError(w, errors.NewAppError(errors.ErrCodeInternal, "Failed to fetch insights", err.Error()), r)
		return
	}
	defer rows.Close()

	insights := []models.Insight{}
	for rows.Next() {
		var insight models.Insight
		var inputContextJSON, resultJSON string
		if err := rows.Scan(&insight.ID, &insight.CompanyID, &insight.Type, &inputContextJSON, &resultJSON, &insight.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(inputContextJSON), &insight.InputContext)
		json.Unmarshal([]byte(resultJSON), &insight.Result)
		insights = append(insights, insight)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"insights": insights,
	})
}
