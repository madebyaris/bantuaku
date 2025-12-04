package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bantuaku/backend/logger"
	"github.com/bantuaku/backend/services/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeService struct {
	secretKey      string
	webhookSecret  string
	db             *storage.Postgres
	log            logger.Logger
}

func NewStripeService(secretKey, webhookSecret string, db *storage.Postgres) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		db:            db,
		log:           *logger.Default(),
	}
}

// CreateCheckoutSession creates a Stripe checkout session for subscription
type CreateCheckoutSessionRequest struct {
	CompanyID string `json:"company_id"`
	PlanID    string `json:"plan_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type CheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

func (s *StripeService) CreateCheckoutSession(ctx context.Context, req CreateCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	// Get plan details
	var planName, stripePriceID string
	var priceMonthly float64
	err := s.db.Pool().QueryRow(ctx, `
		SELECT name, stripe_price_id_monthly, price_monthly
		FROM subscription_plans
		WHERE id = $1 AND is_active = true
	`, req.PlanID).Scan(&planName, &stripePriceID, &priceMonthly)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Get company details
	var companyName, ownerEmail string
	err = s.db.Pool().QueryRow(ctx, `
		SELECT c.name, u.email
		FROM companies c
		JOIN users u ON c.owner_user_id = u.id
		WHERE c.id = $1
	`, req.CompanyID).Scan(&companyName, &ownerEmail)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	// Get or create Stripe customer
	var stripeCustomerID string
	err = s.db.Pool().QueryRow(ctx, `
		SELECT stripe_customer_id
		FROM subscriptions
		WHERE company_id = $1 AND stripe_customer_id IS NOT NULL
		LIMIT 1
	`, req.CompanyID).Scan(&stripeCustomerID)

	if err != nil {
		// Create new Stripe customer
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(ownerEmail),
			Metadata: map[string]string{
				"company_id": req.CompanyID,
				"company_name": companyName,
			},
		}
		cust, err := customer.New(customerParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		stripeCustomerID = cust.ID
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Metadata: map[string]string{
			"company_id": req.CompanyID,
			"plan_id":    req.PlanID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"company_id": req.CompanyID,
				"plan_id":    req.PlanID,
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

// ProcessWebhook processes Stripe webhook events
func (s *StripeService) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Check if event already processed (idempotency)
	var alreadyProcessed bool
	err = s.db.Pool().QueryRow(ctx, `
		SELECT processed FROM stripe_webhooks WHERE stripe_event_id = $1
	`, event.ID).Scan(&alreadyProcessed)
	if err == nil && alreadyProcessed {
		s.log.Info("Webhook event already processed", "event_id", event.ID)
		return nil
	}

	// Store webhook event
	webhookID := uuid.New().String()
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO stripe_webhooks (id, stripe_event_id, event_type, payload, processed, created_at)
		VALUES ($1, $2, $3, $4, false, NOW())
	`, webhookID, event.ID, event.Type, string(payload))
	if err != nil {
		s.log.Error("Failed to store webhook event", "error", err)
		// Continue processing even if storage fails
	}

	// Process event based on type
	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.created", "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_succeeded":
		return s.handlePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(ctx, event)
	default:
		s.log.Info("Unhandled webhook event type", "type", event.Type)
	}

	// Mark as processed
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE stripe_webhooks
		SET processed = true, processed_at = NOW()
		WHERE stripe_event_id = $1
	`, event.ID)
	if err != nil {
		s.log.Error("Failed to mark webhook as processed", "error", err)
	}

	return nil
}

func (s *StripeService) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	companyID := sess.Metadata["company_id"]
	planID := sess.Metadata["plan_id"]

	if companyID == "" || planID == "" {
		return errors.New("missing company_id or plan_id in checkout session metadata")
	}

	// Get subscription from Stripe
	subscriptionID := sess.Subscription.ID
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Create or update subscription in database
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO subscriptions (
			id, company_id, plan_id, status, stripe_subscription_id,
			stripe_customer_id, current_period_start, current_period_end,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (stripe_subscription_id) DO UPDATE SET
			status = $4,
			current_period_start = $7,
			current_period_end = $8,
			updated_at = NOW()
	`, uuid.New().String(), companyID, planID, sub.Status, sub.ID, sub.Customer.ID,
		time.Unix(sub.CurrentPeriodStart, 0), time.Unix(sub.CurrentPeriodEnd, 0))
	if err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	// Update company subscription plan
	_, err = s.db.Pool().Exec(ctx, `
		UPDATE companies
		SET subscription_plan = (SELECT name FROM subscription_plans WHERE id = $1)
		WHERE id = $2
	`, planID, companyID)
	if err != nil {
		s.log.Error("Failed to update company subscription plan", "error", err)
	}

	return nil
}

func (s *StripeService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Update subscription in database
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = $1,
		    current_period_start = $2,
		    current_period_end = $3,
		    cancel_at_period_end = $4,
		    updated_at = NOW()
		WHERE stripe_subscription_id = $5
	`, sub.Status, time.Unix(sub.CurrentPeriodStart, 0), time.Unix(sub.CurrentPeriodEnd, 0),
		sub.CancelAtPeriodEnd, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *StripeService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Update subscription status to canceled
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = 'canceled',
		    canceled_at = NOW(),
		    updated_at = NOW()
		WHERE stripe_subscription_id = $1
	`, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update company subscription plan to free
	var companyID string
	err = s.db.Pool().QueryRow(ctx, `
		SELECT company_id FROM subscriptions WHERE stripe_subscription_id = $1
	`, sub.ID).Scan(&companyID)
	if err == nil {
		_, err = s.db.Pool().Exec(ctx, `
			UPDATE companies SET subscription_plan = 'free' WHERE id = $1
		`, companyID)
		if err != nil {
			s.log.Error("Failed to update company subscription plan", "error", err)
		}
	}

	return nil
}

func (s *StripeService) handlePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	if invoice.Subscription == nil {
		return nil // Not a subscription invoice
	}

	// Get subscription
	var subscriptionID, companyID string
	err := s.db.Pool().QueryRow(ctx, `
		SELECT id, company_id FROM subscriptions WHERE stripe_subscription_id = $1
	`, invoice.Subscription.ID).Scan(&subscriptionID, &companyID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Record payment
	paymentID := uuid.New().String()
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO payments (
			id, subscription_id, company_id, amount, currency, status,
			stripe_payment_intent_id, stripe_invoice_id, payment_method,
			paid_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, 'succeeded', $6, $7, 'card', NOW(), NOW())
	`, paymentID, subscriptionID, companyID,
		float64(invoice.AmountPaid)/100, // Convert from cents
		string(invoice.Currency),
		invoice.PaymentIntent.ID,
		invoice.ID)
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}

	return nil
}

func (s *StripeService) handlePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	if invoice.Subscription == nil {
		return nil
	}

	// Update subscription status to past_due
	_, err := s.db.Pool().Exec(ctx, `
		UPDATE subscriptions
		SET status = 'past_due', updated_at = NOW()
		WHERE stripe_subscription_id = $1
	`, invoice.Subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

