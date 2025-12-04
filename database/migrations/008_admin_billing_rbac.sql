-- Bantuaku - Admin, Billing, and RBAC Tables
-- Migration 008: Admin Panel and Subscription Management
-- PostgreSQL 18
-- Dependencies: 001_init_schema.sql (users table), 003_add_chat_tables.sql (companies table)

-- User roles (extend users table)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user';
-- Roles: 'user', 'admin', 'super_admin'

-- Subscription plans
CREATE TABLE IF NOT EXISTS subscription_plans (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,  -- 'free', 'pro', 'enterprise'
    display_name VARCHAR(255) NOT NULL,
    price_monthly NUMERIC(12, 2) NOT NULL,  -- IDR
    price_yearly NUMERIC(12, 2),  -- Optional yearly pricing
    currency VARCHAR(3) DEFAULT 'IDR',
    max_stores INT,  -- NULL = unlimited
    max_products INT,  -- NULL = unlimited
    features JSONB NOT NULL,  -- Feature flags
    stripe_price_id_monthly VARCHAR(255),  -- Stripe Price ID
    stripe_price_id_yearly VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_id VARCHAR(36) NOT NULL REFERENCES subscription_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'canceled', 'past_due', 'trialing'
    stripe_subscription_id VARCHAR(255) UNIQUE,  -- Stripe Subscription ID
    stripe_customer_id VARCHAR(255),  -- Stripe Customer ID
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT false,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payments (transaction history)
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) PRIMARY KEY,
    subscription_id VARCHAR(36) NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    status VARCHAR(20) NOT NULL,  -- 'pending', 'succeeded', 'failed', 'refunded'
    stripe_payment_intent_id VARCHAR(255) UNIQUE,
    stripe_invoice_id VARCHAR(255),
    payment_method VARCHAR(50),  -- 'card', 'bank_transfer', etc.
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Stripe webhooks (for idempotency)
CREATE TABLE IF NOT EXISTS stripe_webhooks (
    id VARCHAR(36) PRIMARY KEY,
    stripe_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT false,
    processed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit logs (security and compliance)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(36) REFERENCES users(id),
    company_id VARCHAR(36) REFERENCES companies(id),
    action VARCHAR(100) NOT NULL,  -- 'user.created', 'subscription.activated', etc.
    resource_type VARCHAR(50),  -- 'user', 'subscription', 'payment', etc.
    resource_id VARCHAR(36),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,  -- Additional context
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_active ON subscription_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_name ON subscription_plans(name);

CREATE INDEX IF NOT EXISTS idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription_id ON subscriptions(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_period_end ON subscriptions(current_period_end);

CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payments_company_id ON payments(company_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_stripe_payment_intent_id ON payments(stripe_payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_stripe_webhooks_event_id ON stripe_webhooks(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_stripe_webhooks_processed ON stripe_webhooks(processed, created_at);
CREATE INDEX IF NOT EXISTS idx_stripe_webhooks_event_type ON stripe_webhooks(event_type);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_company_id ON audit_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- Seed default subscription plans
INSERT INTO subscription_plans (id, name, display_name, price_monthly, max_stores, max_products, features) VALUES
    ('free-plan', 'free', 'Free Plan', 0, 1, 10, '{"forecast": true, "ai_chat": false, "trends": false, "sentiment": false}'::jsonb),
    ('pro-plan', 'pro', 'Pro Plan', 500000, 1, NULL, '{"forecast": true, "ai_chat": true, "trends": true, "sentiment": true}'::jsonb),
    ('enterprise-plan', 'enterprise', 'Enterprise Plan', 0, NULL, NULL, '{"forecast": true, "ai_chat": true, "trends": true, "sentiment": true, "multi_store": true, "custom_integrations": true}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- Comments
COMMENT ON TABLE subscription_plans IS 'Available subscription plans (free, pro, enterprise)';
COMMENT ON TABLE subscriptions IS 'Active subscriptions per company';
COMMENT ON TABLE payments IS 'Payment transaction history from Stripe';
COMMENT ON TABLE stripe_webhooks IS 'Stripe webhook events for idempotency and event tracking';
COMMENT ON TABLE audit_logs IS 'Security audit trail for sensitive actions (user management, subscriptions, etc.)';
COMMENT ON COLUMN users.role IS 'User role: user (default), admin, super_admin';

