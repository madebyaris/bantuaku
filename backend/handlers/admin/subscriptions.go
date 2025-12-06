package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/audit"
	"github.com/bantuaku/backend/services/transactions"
	"github.com/google/uuid"
)

// SubscriptionStats represents subscription statistics
type SubscriptionStats struct {
	TotalSubscriptions  int             `json:"total_subscriptions"`
	ActiveSubscriptions int             `json:"active_subscriptions"`
	TrialingCount       int             `json:"trialing_count"`
	CanceledCount       int             `json:"canceled_count"`
	PastDueCount        int             `json:"past_due_count"`
	MRR                 float64         `json:"mrr"` // Monthly Recurring Revenue
	PlanBreakdown       []PlanBreakdown `json:"plan_breakdown"`
}

// PlanBreakdown shows subscriptions per plan
type PlanBreakdown struct {
	PlanID       string  `json:"plan_id"`
	PlanName     string  `json:"plan_name"`
	Count        int     `json:"count"`
	PriceMonthly float64 `json:"price_monthly"`
}

// GetSubscriptionStats returns subscription statistics
func (h *AdminHandler) GetSubscriptionStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var stats SubscriptionStats

	// Total subscriptions
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&stats.TotalSubscriptions)

	// Active subscriptions
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'active'`).Scan(&stats.ActiveSubscriptions)

	// Trialing
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'trialing'`).Scan(&stats.TrialingCount)

	// Canceled
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'canceled'`).Scan(&stats.CanceledCount)

	// Past due
	h.db.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status = 'past_due'`).Scan(&stats.PastDueCount)

	// MRR calculation - sum of active subscription plan prices
	h.db.Pool().QueryRow(ctx, `
		SELECT COALESCE(SUM(p.price_monthly), 0) 
		FROM subscriptions s 
		JOIN subscription_plans p ON s.plan_id = p.id 
		WHERE s.status = 'active'
	`).Scan(&stats.MRR)

	// Plan breakdown
	rows, err := h.db.Pool().Query(ctx, `
		SELECT p.id, p.display_name, COUNT(s.id), p.price_monthly
		FROM subscription_plans p
		LEFT JOIN subscriptions s ON s.plan_id = p.id AND s.status = 'active'
		WHERE p.is_active = true
		GROUP BY p.id, p.display_name, p.price_monthly
		ORDER BY p.price_monthly ASC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pb PlanBreakdown
			if err := rows.Scan(&pb.PlanID, &pb.PlanName, &pb.Count, &pb.PriceMonthly); err == nil {
				stats.PlanBreakdown = append(stats.PlanBreakdown, pb)
			}
		}
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// Subscription represents a subscription in admin context
type Subscription struct {
	ID                   string     `json:"id"`
	CompanyID            string     `json:"company_id"`
	CompanyName          string     `json:"company_name"`
	OwnerEmail           string     `json:"owner_email"`
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
			COALESCE(u.email, '') as owner_email,
			s.plan_id, p.name as plan_name,
			s.status, s.stripe_subscription_id, s.stripe_customer_id,
			s.current_period_start, s.current_period_end,
			s.cancel_at_period_end, s.canceled_at,
			s.created_at, s.updated_at
		FROM subscriptions s
		JOIN companies c ON s.company_id = c.id
		JOIN subscription_plans p ON s.plan_id = p.id
		LEFT JOIN users u ON c.owner_user_id = u.id
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
			&s.ID, &s.CompanyID, &s.CompanyName, &s.OwnerEmail,
			&s.PlanID, &s.PlanName,
			&s.Status, &s.StripeSubscriptionID, &s.StripeCustomerID,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&s.CancelAtPeriodEnd, &canceledAt,
			&s.CreatedAt, &updatedAt,
		); err != nil {
			logger.Error("Failed to scan subscription", "error", err)
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
	path := r.URL.Path
	subscriptionID := path[len("/api/v1/admin/subscriptions/"):]

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

func (h *AdminHandler) UpdateSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Extract subscription ID from path (format: /api/v1/admin/subscriptions/{id}/status)
	path := r.URL.Path
	subscriptionID := path[len("/api/v1/admin/subscriptions/"):]
	if idx := len(subscriptionID) - len("/status"); idx > 0 && subscriptionID[idx:] == "/status" {
		subscriptionID = subscriptionID[:idx]
	}

	var req UpdateSubscriptionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	// Get old status and company_id before updating
	var oldStatus, companyID string
	err := h.db.Pool().QueryRow(ctx, `
		SELECT status, company_id FROM subscriptions WHERE id = $1
	`, subscriptionID).Scan(&oldStatus, &companyID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "get subscription")
		h.respondError(w, appErr, r)
		return
	}

	_, err = h.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, req.Status, subscriptionID)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "update subscription status")
		h.respondError(w, appErr, r)
		return
	}

	// Log transaction
	userID := middleware.GetUserID(ctx)
	var changedByUserID *string
	if userID != "" {
		changedByUserID = &userID
	}
	transactionService := transactions.NewService(h.db)
	oldStatusPtr := &oldStatus
	newStatusPtr := &req.Status
	if err := transactionService.CreateTransaction(ctx, subscriptionID, companyID, transactions.EventTypeStatusChange, nil, nil, oldStatusPtr, newStatusPtr, changedByUserID, map[string]interface{}{
		"reason": "admin_update",
	}); err != nil {
		// Log error but don't fail the request - use logger directly since AdminHandler log may not be initialized
		logger.Warn("Failed to log subscription transaction", "error", err)
	}

	// Log audit event
	if h.auditLogger != nil {
		h.auditLogger.LogResourceAction(ctx, r, audit.ActionSubscriptionUpdated, "subscription", subscriptionID, map[string]interface{}{
			"old_status": oldStatus,
			"new_status": req.Status,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Subscription status updated successfully"})
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

	// Log transaction
	userID := middleware.GetUserID(ctx)
	var changedByUserID *string
	if userID != "" {
		changedByUserID = &userID
	}
	transactionService := transactions.NewService(h.db)
	newPlanID := &req.PlanID
	newStatus := "active"
	if err := transactionService.CreateTransaction(ctx, subscriptionID, req.CompanyID, transactions.EventTypeCreate, nil, newPlanID, nil, &newStatus, changedByUserID, map[string]interface{}{
		"period_start": req.CurrentPeriodStart,
		"period_end":   req.CurrentPeriodEnd,
	}); err != nil {
		// Log error but don't fail the request - use logger directly since AdminHandler log may not be initialized
		logger.Warn("Failed to log subscription transaction", "error", err)
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
