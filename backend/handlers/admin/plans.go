package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/services/audit"
	"github.com/google/uuid"
)

// SubscriptionPlan represents a subscription plan in admin context
type SubscriptionPlan struct {
	ID                           string                 `json:"id"`
	Name                         string                 `json:"name"`
	DisplayName                  string                 `json:"display_name"`
	PriceMonthly                 float64                `json:"price_monthly"`
	PriceYearly                  *float64               `json:"price_yearly,omitempty"`
	Currency                     string                 `json:"currency"`
	MaxStores                    *int                   `json:"max_stores,omitempty"`
	MaxProducts                  *int                   `json:"max_products,omitempty"`
	MaxChatsPerMonth             *int                   `json:"max_chats_per_month,omitempty"`
	MaxFileUploadsPerMonth       *int                   `json:"max_file_uploads_per_month,omitempty"`
	MaxFileSizeMB                *int                   `json:"max_file_size_mb,omitempty"`
	MaxForecastRefreshesPerMonth *int                   `json:"max_forecast_refreshes_per_month,omitempty"`
	Features                     map[string]interface{} `json:"features"`
	IsActive                     bool                   `json:"is_active"`
	CreatedAt                    time.Time              `json:"created_at"`
	UpdatedAt                    *time.Time             `json:"updated_at,omitempty"`
}

// ListPlans lists all subscription plans with pagination
func (h *AdminHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	rows, err := h.db.Pool().Query(ctx, `
		SELECT 
			id, name, display_name, 
			price_monthly, price_yearly, currency,
			max_stores, max_products,
			max_chats_per_month, max_file_uploads_per_month, 
			max_file_size_mb, max_forecast_refreshes_per_month,
			features, is_active, created_at, updated_at
		FROM subscription_plans
		ORDER BY 
			CASE name 
				WHEN 'free' THEN 1 
				WHEN 'pro' THEN 2 
				WHEN 'enterprise' THEN 3 
				ELSE 4 
			END,
			created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "list subscription plans")
		h.respondError(w, appErr, r)
		return
	}
	defer rows.Close()

	var plans []SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		var priceYearly *float64
		var maxStores, maxProducts, maxChats, maxUploads, maxFileSize, maxForecasts *int
		var featuresJSON []byte
		var updatedAt *time.Time

		if err := rows.Scan(
			&p.ID, &p.Name, &p.DisplayName,
			&p.PriceMonthly, &priceYearly, &p.Currency,
			&maxStores, &maxProducts,
			&maxChats, &maxUploads, &maxFileSize, &maxForecasts,
			&featuresJSON, &p.IsActive, &p.CreatedAt, &updatedAt,
		); err != nil {
			h.log.Error("Failed to scan plan", "error", err)
			continue
		}

		p.PriceYearly = priceYearly
		p.MaxStores = maxStores
		p.MaxProducts = maxProducts
		p.MaxChatsPerMonth = maxChats
		p.MaxFileUploadsPerMonth = maxUploads
		p.MaxFileSizeMB = maxFileSize
		p.MaxForecastRefreshesPerMonth = maxForecasts
		p.UpdatedAt = updatedAt

		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
				h.log.Error("Failed to unmarshal features", "error", err)
				p.Features = make(map[string]interface{})
			}
		}

		plans = append(plans, p)
	}

	// Get total count
	var total int
	err = h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscription_plans`).Scan(&total)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "count subscription plans")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetPlan retrieves a single subscription plan by ID
func (h *AdminHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planID := r.PathValue("id")
	if planID == "" {
		appErr := errors.NewValidationError("Plan ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	var p SubscriptionPlan
	var priceYearly *float64
	var maxStores, maxProducts, maxChats, maxUploads, maxFileSize, maxForecasts *int
	var featuresJSON []byte
	var updatedAt *time.Time

	err := h.db.Pool().QueryRow(ctx, `
		SELECT 
			id, name, display_name, 
			price_monthly, price_yearly, currency,
			max_stores, max_products,
			max_chats_per_month, max_file_uploads_per_month, 
			max_file_size_mb, max_forecast_refreshes_per_month,
			features, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`, planID).Scan(
		&p.ID, &p.Name, &p.DisplayName,
		&p.PriceMonthly, &priceYearly, &p.Currency,
		&maxStores, &maxProducts,
		&maxChats, &maxUploads, &maxFileSize, &maxForecasts,
		&featuresJSON, &p.IsActive, &p.CreatedAt, &updatedAt,
	)
	if err != nil {
		appErr := errors.NewNotFoundError("Plan not found")
		h.respondError(w, appErr, r)
		return
	}

	p.PriceYearly = priceYearly
	p.MaxStores = maxStores
	p.MaxProducts = maxProducts
	p.MaxChatsPerMonth = maxChats
	p.MaxFileUploadsPerMonth = maxUploads
	p.MaxFileSizeMB = maxFileSize
	p.MaxForecastRefreshesPerMonth = maxForecasts
	p.UpdatedAt = updatedAt

	if len(featuresJSON) > 0 {
		json.Unmarshal(featuresJSON, &p.Features)
	}

	h.respondJSON(w, http.StatusOK, p)
}

// CreatePlanRequest represents the request body for creating a plan
type CreatePlanRequest struct {
	Name                         string                 `json:"name"`
	DisplayName                  string                 `json:"display_name"`
	PriceMonthly                 float64                `json:"price_monthly"`
	PriceYearly                  *float64               `json:"price_yearly,omitempty"`
	Currency                     string                 `json:"currency"`
	MaxStores                    *int                   `json:"max_stores,omitempty"`
	MaxProducts                  *int                   `json:"max_products,omitempty"`
	MaxChatsPerMonth             *int                   `json:"max_chats_per_month,omitempty"`
	MaxFileUploadsPerMonth       *int                   `json:"max_file_uploads_per_month,omitempty"`
	MaxFileSizeMB                *int                   `json:"max_file_size_mb,omitempty"`
	MaxForecastRefreshesPerMonth *int                   `json:"max_forecast_refreshes_per_month,omitempty"`
	Features                     map[string]interface{} `json:"features"`
}

// CreatePlan creates a new subscription plan
func (h *AdminHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	if req.Name == "" {
		appErr := errors.NewValidationError("Plan name is required", "")
		h.respondError(w, appErr, r)
		return
	}
	if req.DisplayName == "" {
		appErr := errors.NewValidationError("Display name is required", "")
		h.respondError(w, appErr, r)
		return
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}

	// Check if name already exists
	var existingID string
	err := h.db.Pool().QueryRow(ctx, `SELECT id FROM subscription_plans WHERE name = $1`, req.Name).Scan(&existingID)
	if err == nil {
		appErr := errors.NewConflictError("Plan with this name already exists", "")
		h.respondError(w, appErr, r)
		return
	}

	featuresJSON, _ := json.Marshal(req.Features)
	if req.Features == nil {
		featuresJSON = []byte("{}")
	}

	planID := uuid.New().String()
	_, err = h.db.Pool().Exec(ctx, `
		INSERT INTO subscription_plans (
			id, name, display_name, 
			price_monthly, price_yearly, currency,
			max_stores, max_products,
			max_chats_per_month, max_file_uploads_per_month, 
			max_file_size_mb, max_forecast_refreshes_per_month,
			features, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, true, NOW())
	`, planID, req.Name, req.DisplayName,
		req.PriceMonthly, req.PriceYearly, req.Currency,
		req.MaxStores, req.MaxProducts,
		req.MaxChatsPerMonth, req.MaxFileUploadsPerMonth,
		req.MaxFileSizeMB, req.MaxForecastRefreshesPerMonth,
		featuresJSON,
	)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create subscription plan")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionCreated, "subscription_plan", planID, map[string]interface{}{
			"name":         req.Name,
			"display_name": req.DisplayName,
		})
	}

	h.respondJSON(w, http.StatusCreated, map[string]string{
		"id":   planID,
		"name": req.Name,
	})
}

// UpdatePlanRequest represents the request body for updating a plan
type UpdatePlanRequest struct {
	DisplayName                  string                 `json:"display_name"`
	PriceMonthly                 float64                `json:"price_monthly"`
	PriceYearly                  *float64               `json:"price_yearly,omitempty"`
	Currency                     string                 `json:"currency"`
	MaxStores                    *int                   `json:"max_stores,omitempty"`
	MaxProducts                  *int                   `json:"max_products,omitempty"`
	MaxChatsPerMonth             *int                   `json:"max_chats_per_month,omitempty"`
	MaxFileUploadsPerMonth       *int                   `json:"max_file_uploads_per_month,omitempty"`
	MaxFileSizeMB                *int                   `json:"max_file_size_mb,omitempty"`
	MaxForecastRefreshesPerMonth *int                   `json:"max_forecast_refreshes_per_month,omitempty"`
	Features                     map[string]interface{} `json:"features"`
	IsActive                     *bool                  `json:"is_active,omitempty"`
}

// UpdatePlan updates an existing subscription plan
func (h *AdminHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planID := r.PathValue("id")
	if planID == "" {
		appErr := errors.NewValidationError("Plan ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	if req.DisplayName == "" {
		appErr := errors.NewValidationError("Display name is required", "")
		h.respondError(w, appErr, r)
		return
	}

	featuresJSON, _ := json.Marshal(req.Features)
	if req.Features == nil {
		featuresJSON = []byte("{}")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	_, err := h.db.Pool().Exec(ctx, `
		UPDATE subscription_plans SET
			display_name = $1,
			price_monthly = $2,
			price_yearly = $3,
			currency = $4,
			max_stores = $5,
			max_products = $6,
			max_chats_per_month = $7,
			max_file_uploads_per_month = $8,
			max_file_size_mb = $9,
			max_forecast_refreshes_per_month = $10,
			features = $11,
			is_active = $12,
			updated_at = NOW()
		WHERE id = $13
	`, req.DisplayName, req.PriceMonthly, req.PriceYearly, req.Currency,
		req.MaxStores, req.MaxProducts,
		req.MaxChatsPerMonth, req.MaxFileUploadsPerMonth,
		req.MaxFileSizeMB, req.MaxForecastRefreshesPerMonth,
		featuresJSON, isActive, planID,
	)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update subscription plan")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionUpdated, "subscription_plan", planID, map[string]interface{}{
			"display_name": req.DisplayName,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Plan updated successfully"})
}

// DeletePlan soft-deletes a subscription plan by setting is_active to false
func (h *AdminHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planID := r.PathValue("id")
	if planID == "" {
		appErr := errors.NewValidationError("Plan ID is required", "")
		h.respondError(w, appErr, r)
		return
	}

	// Check if plan has active subscriptions
	var activeCount int
	err := h.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM subscriptions WHERE plan_id = $1 AND status = 'active'
	`, planID).Scan(&activeCount)
	if err == nil && activeCount > 0 {
		appErr := errors.NewConflictError("Cannot delete plan with active subscriptions", "")
		h.respondError(w, appErr, r)
		return
	}

	// Soft delete by setting is_active to false
	_, err = h.db.Pool().Exec(ctx, `
		UPDATE subscription_plans SET is_active = false, updated_at = NOW() WHERE id = $1
	`, planID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "delete subscription plan")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, "subscription_plan.deleted", "subscription_plan", planID, nil)
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Plan deactivated successfully"})
}
