package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/services/audit"
	"github.com/google/uuid"
)

// Subscription represents a subscription in admin context
type Subscription struct {
	ID                   string     `json:"id"`
	CompanyID            string     `json:"company_id"`
	CompanyName          string     `json:"company_name"`
	PlanID               string     `json:"plan_id"`
	PlanName             string     `json:"plan_name"`
	Status               string     `json:"status"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	CurrentPeriodStart   time.Time  `json:"current_period_start"`
	CurrentPeriodEnd     time.Time  `json:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
}

// ListSubscriptions lists all subscriptions with pagination
func (h *AdminHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
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
			s.id, s.company_id, c.name as company_name,
			s.plan_id, p.name as plan_name,
			s.status, s.stripe_subscription_id, s.stripe_customer_id,
			s.current_period_start, s.current_period_end,
			s.cancel_at_period_end, s.canceled_at,
			s.created_at, s.updated_at
		FROM subscriptions s
		JOIN companies c ON s.company_id = c.id
		JOIN subscription_plans p ON s.plan_id = p.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "list subscriptions")
		h.respondError(w, appErr, r)
		return
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var s Subscription
		var canceledAt, updatedAt *time.Time
		if err := rows.Scan(
			&s.ID, &s.CompanyID, &s.CompanyName,
			&s.PlanID, &s.PlanName,
			&s.Status, &s.StripeSubscriptionID, &s.StripeCustomerID,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&s.CancelAtPeriodEnd, &canceledAt,
			&s.CreatedAt, &updatedAt,
		); err != nil {
			h.log.Error("Failed to scan subscription", "error", err)
			continue
		}
		if canceledAt != nil {
			s.CanceledAt = canceledAt
		}
		if updatedAt != nil {
			s.UpdatedAt = updatedAt
		}
		subscriptions = append(subscriptions, s)
	}

	// Get total count
	var total int
	err = h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&total)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "count subscriptions")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetSubscription retrieves a single subscription by ID
func (h *AdminHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract subscription ID from path
	subscriptionID := r.PathValue("id")

	var s Subscription
	var canceledAt, updatedAt *time.Time
	err := h.db.Pool().QueryRow(ctx, `
		SELECT 
			s.id, s.company_id, c.name as company_name,
			s.plan_id, p.name as plan_name,
			s.status, s.stripe_subscription_id, s.stripe_customer_id,
			s.current_period_start, s.current_period_end,
			s.cancel_at_period_end, s.canceled_at,
			s.created_at, s.updated_at
		FROM subscriptions s
		JOIN companies c ON s.company_id = c.id
		JOIN subscription_plans p ON s.plan_id = p.id
		WHERE s.id = $1
	`, subscriptionID).Scan(
		&s.ID, &s.CompanyID, &s.CompanyName,
		&s.PlanID, &s.PlanName,
		&s.Status, &s.StripeSubscriptionID, &s.StripeCustomerID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.CancelAtPeriodEnd, &canceledAt,
		&s.CreatedAt, &updatedAt,
	)
	if err != nil {
		appErr := errors.NewNotFoundError("Subscription not found")
		h.respondError(w, appErr, r)
		return
	}

	if canceledAt != nil {
		s.CanceledAt = canceledAt
	}
	if updatedAt != nil {
		s.UpdatedAt = updatedAt
	}

	h.respondJSON(w, http.StatusOK, s)
}

// UpdateSubscriptionStatus updates a subscription's status
type UpdateSubscriptionStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active canceled past_due trialing"`
}

// UpdateSubscriptionPlanRequest updates plan (and optionally status)
type UpdateSubscriptionPlanRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active canceled past_due trialing"`
}

func (h *AdminHandler) UpdateSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract subscription ID from path (format: /api/v1/admin/subscriptions/{id}/status)
	subscriptionID := r.PathValue("id")

	var req UpdateSubscriptionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	_, err := h.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, req.Status, subscriptionID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update subscription status")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionUpdated, "subscription", subscriptionID, map[string]interface{}{
			"new_status": req.Status,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Subscription status updated successfully"})
}

// UpdateSubscriptionPlan updates a subscription's plan (and optional status) and syncs company.subscription_plan
func (h *AdminHandler) UpdateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	subscriptionID := r.PathValue("id")

	var req UpdateSubscriptionPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	// Validate plan exists and get plan name
	var planName string
	err := h.db.Pool().QueryRow(ctx, `SELECT name FROM subscription_plans WHERE id = $1 OR name = $1 LIMIT 1`, req.PlanID).Scan(&planName)
	if err != nil {
		appErr := errors.NewNotFoundError("Plan not found")
		h.respondError(w, appErr, r)
		return
	}

	// Get company_id for this subscription
	var companyID string
	err = h.db.Pool().QueryRow(ctx, `SELECT company_id FROM subscriptions WHERE id = $1`, subscriptionID).Scan(&companyID)
	if err != nil {
		appErr := errors.NewNotFoundError("Subscription not found")
		h.respondError(w, appErr, r)
		return
	}

	// Update subscription plan (and status if provided)
	if strings.TrimSpace(req.Status) != "" {
		if _, err := h.db.Pool().Exec(ctx, `
			UPDATE subscriptions
			SET plan_id = $1, status = $2, updated_at = NOW()
			WHERE id = $3
		`, req.PlanID, req.Status, subscriptionID); err != nil {
			appErr := errors.NewDatabaseError(err, "update subscription plan/status")
			h.respondError(w, appErr, r)
			return
		}
	} else {
		if _, err := h.db.Pool().Exec(ctx, `
			UPDATE subscriptions
			SET plan_id = $1, updated_at = NOW()
			WHERE id = $2
		`, req.PlanID, subscriptionID); err != nil {
			appErr := errors.NewDatabaseError(err, "update subscription plan")
			h.respondError(w, appErr, r)
			return
		}
	}

	// Sync company subscription_plan
	_, _ = h.db.Pool().Exec(ctx, `UPDATE companies SET subscription_plan = $1 WHERE id = $2`, planName, companyID)

	// Audit log
	if h.auditLogger != nil {
		meta := map[string]interface{}{
			"new_plan_id":   req.PlanID,
			"new_plan_name": planName,
		}
		if strings.TrimSpace(req.Status) != "" {
			meta["new_status"] = req.Status
		}
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionUpdated, "subscription", subscriptionID, meta)
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Subscription plan updated successfully"})
}

// CreateSubscription creates a new subscription
type CreateSubscriptionRequest struct {
	CompanyID          string    `json:"company_id" validate:"required"`
	PlanID             string    `json:"plan_id" validate:"required"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
}

func (h *AdminHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	// Set default period if not provided
	if req.CurrentPeriodStart.IsZero() {
		req.CurrentPeriodStart = time.Now()
	}
	if req.CurrentPeriodEnd.IsZero() {
		req.CurrentPeriodEnd = req.CurrentPeriodStart.AddDate(0, 1, 0) // 1 month
	}

	subscriptionID := uuid.New().String()
	_, err := h.db.Pool().Exec(ctx, `
		INSERT INTO subscriptions (id, company_id, plan_id, status, current_period_start, current_period_end, created_at)
		VALUES ($1, $2, $3, 'active', $4, $5, NOW())
	`, subscriptionID, req.CompanyID, req.PlanID, req.CurrentPeriodStart, req.CurrentPeriodEnd)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "create subscription")
		h.respondError(w, appErr, r)
		return
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionCreated, "subscription", subscriptionID, map[string]interface{}{
			"company_id": req.CompanyID,
			"plan_id":    req.PlanID,
		})
	}

	h.respondJSON(w, http.StatusCreated, map[string]string{
		"id":     subscriptionID,
		"status": "active",
	})
}
