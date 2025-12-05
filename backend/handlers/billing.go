package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bantuaku/backend/errors"
	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/middleware"
	"github.com/bantuaku/backend/services/billing"
	"github.com/bantuaku/backend/services/storage"
)

type BillingHandler struct {
	stripeService *billing.StripeService
	db            *storage.Postgres
	log           logger.Logger
}

func NewBillingHandler(stripeService *billing.StripeService, db *storage.Postgres) *BillingHandler {
	return &BillingHandler{
		stripeService: stripeService,
		db:            db,
		log:           *logger.Default(),
	}
}

// CreateCheckoutSessionRequest represents a request to create a checkout session
type CreateCheckoutSessionRequest struct {
	PlanID     string `json:"plan_id" validate:"required"`
	SuccessURL string `json:"success_url" validate:"required,url"`
	CancelURL  string `json:"cancel_url" validate:"required,url"`
}

// Helper methods
func (h *BillingHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *BillingHandler) respondError(w http.ResponseWriter, err error, r *http.Request) {
	errors.WriteJSONError(w, err, errors.GetErrorCode(err))
}

func (h *BillingHandler) parseJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.NewValidationError("Invalid JSON format", err.Error())
	}
	return nil
}

// CreateCheckoutSession creates a Stripe checkout session
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		appErr := errors.NewUnauthorizedError("Company ID not found in token")
		h.respondError(w, appErr, r)
		return
	}

	var req CreateCheckoutSessionRequest
	if err := h.parseJSON(r, &req); err != nil {
		h.respondError(w, err, r)
		return
	}

	session, err := h.stripeService.CreateCheckoutSession(ctx, billing.CreateCheckoutSessionRequest{
		CompanyID:  companyID,
		PlanID:     req.PlanID,
		SuccessURL: req.SuccessURL,
		CancelURL:  req.CancelURL,
	})
	if err != nil {
		appErr := errors.NewInternalError(err, "Failed to create checkout session")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, session)
}

// HandleWebhook processes Stripe webhook events
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read raw body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		appErr := errors.NewValidationError("Failed to read request body", err.Error())
		h.respondError(w, appErr, r)
		return
	}

	// Get Stripe signature header
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		appErr := errors.NewUnauthorizedError("Missing Stripe-Signature header")
		h.respondError(w, appErr, r)
		return
	}

	// Process webhook
	if err := h.stripeService.ProcessWebhook(ctx, payload, signature); err != nil {
		h.log.Error("Webhook processing failed", "error", err)
		appErr := errors.NewInternalError(err, "Failed to process webhook")
		h.respondError(w, appErr, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

// GetSubscription retrieves the current subscription for the company
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := middleware.GetCompanyID(ctx)
	if companyID == "" {
		appErr := errors.NewUnauthorizedError("Company ID not found in token")
		h.respondError(w, appErr, r)
		return
	}

	var subscriptionID, planID, status, stripeSubscriptionID string
	var currentPeriodStart, currentPeriodEnd interface{}
	err := h.db.Pool().QueryRow(ctx, `
		SELECT id, plan_id, status, stripe_subscription_id,
		       current_period_start, current_period_end
		FROM subscriptions
		WHERE company_id = $1 AND status IN ('active', 'trialing')
		ORDER BY created_at DESC
		LIMIT 1
	`, companyID).Scan(&subscriptionID, &planID, &status, &stripeSubscriptionID,
		&currentPeriodStart, &currentPeriodEnd)
	if err != nil {
		appErr := errors.NewNotFoundError("No active subscription found")
		h.respondError(w, appErr, r)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":                    subscriptionID,
		"plan_id":               planID,
		"status":                 status,
		"stripe_subscription_id": stripeSubscriptionID,
		"current_period_start":   currentPeriodStart,
		"current_period_end":     currentPeriodEnd,
	})
}

// ListPlans lists available subscription plans
func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := h.db.Pool().Query(ctx, `
		SELECT id, name, display_name, price_monthly, price_yearly,
		       currency, max_stores, max_products, features, stripe_price_id_monthly
		FROM subscription_plans
		WHERE is_active = true
		ORDER BY price_monthly ASC
	`)
	if err != nil {
		appErr := errors.NewDatabaseError(err, "list plans")
		h.respondError(w, appErr, r)
		return
	}
	defer rows.Close()

	type Plan struct {
		ID                   string                 `json:"id"`
		Name                 string                 `json:"name"`
		DisplayName          string                 `json:"display_name"`
		PriceMonthly         float64                `json:"price_monthly"`
		PriceYearly          *float64               `json:"price_yearly,omitempty"`
		Currency             string                 `json:"currency"`
		MaxStores            *int                   `json:"max_stores,omitempty"`
		MaxProducts          *int                   `json:"max_products,omitempty"`
		Features             map[string]interface{} `json:"features"`
		StripePriceIDMonthly *string                `json:"stripe_price_id_monthly,omitempty"`
	}

	var plans []Plan
	for rows.Next() {
		var p Plan
		var priceYearly *float64
		var maxStores, maxProducts *int
		var featuresJSON []byte
		var stripePriceID *string

		if err := rows.Scan(
			&p.ID, &p.Name, &p.DisplayName, &p.PriceMonthly, &priceYearly,
			&p.Currency, &maxStores, &maxProducts, &featuresJSON, &stripePriceID,
		); err != nil {
			h.log.Error("Failed to scan plan", "error", err)
			continue
		}

		p.PriceYearly = priceYearly
		p.MaxStores = maxStores
		p.MaxProducts = maxProducts
		p.StripePriceIDMonthly = stripePriceID

		if len(featuresJSON) > 0 {
			if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
				h.log.Error("Failed to unmarshal features", "error", err)
			}
		}

		plans = append(plans, p)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

