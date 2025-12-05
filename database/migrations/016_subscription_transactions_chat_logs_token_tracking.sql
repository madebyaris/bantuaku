-- Migration 016: Subscription Transactions, Chat Usage Logs, and Token Tracking
-- Adds tables for tracking subscription events, aggregate chat usage, and AI token consumption
-- Dependencies: 008_admin_billing_rbac.sql (subscriptions, companies), 003_add_chat_tables.sql (messages, conversations)

-- Subscription Transactions (track all subscription events)
CREATE TABLE IF NOT EXISTS subscription_transactions (
    id VARCHAR(36) PRIMARY KEY,
    subscription_id VARCHAR(36) NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,  -- 'create', 'upgrade', 'downgrade', 'cancel', 'renew', 'status_change'
    old_plan_id VARCHAR(36) REFERENCES subscription_plans(id),
    new_plan_id VARCHAR(36) REFERENCES subscription_plans(id),
    old_status VARCHAR(20),  -- Previous subscription status
    new_status VARCHAR(20),  -- New subscription status
    changed_by_user_id VARCHAR(36) REFERENCES users(id),  -- User who made the change (admin or system)
    metadata JSONB,  -- Additional context (reason, notes, etc.)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Chat Usage Logs (aggregate daily/monthly stats - no message content)
CREATE TABLE IF NOT EXISTS chat_usage_logs (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    date DATE NOT NULL,  -- Date for aggregation (YYYY-MM-DD)
    total_messages INT NOT NULL DEFAULT 0,  -- Total messages sent/received
    total_conversations INT NOT NULL DEFAULT 0,  -- Total conversations created
    unique_users INT NOT NULL DEFAULT 0,  -- Unique users who chatted
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, date)  -- One record per company per day
);

-- Token Usage (track AI token consumption per chat completion)
CREATE TABLE IF NOT EXISTS token_usage (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    conversation_id VARCHAR(36) REFERENCES conversations(id) ON DELETE SET NULL,
    message_id VARCHAR(36),  -- Assistant message ID
    model VARCHAR(100) NOT NULL,  -- AI model used (e.g., 'openai/gpt-4o-mini', 'GLM 4.6')
    provider VARCHAR(50) NOT NULL,  -- 'openrouter', 'kolosal'
    prompt_tokens INT NOT NULL DEFAULT 0,  -- Input tokens
    completion_tokens INT NOT NULL DEFAULT 0,  -- Output tokens
    total_tokens INT NOT NULL DEFAULT 0,  -- Total tokens
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for subscription_transactions
CREATE INDEX IF NOT EXISTS idx_subscription_transactions_subscription_id ON subscription_transactions(subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_transactions_company_id ON subscription_transactions(company_id);
CREATE INDEX IF NOT EXISTS idx_subscription_transactions_event_type ON subscription_transactions(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_transactions_created_at ON subscription_transactions(created_at DESC);

-- Indexes for chat_usage_logs
CREATE INDEX IF NOT EXISTS idx_chat_usage_logs_company_id ON chat_usage_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_chat_usage_logs_date ON chat_usage_logs(date DESC);
CREATE INDEX IF NOT EXISTS idx_chat_usage_logs_company_date ON chat_usage_logs(company_id, date DESC);

-- Indexes for token_usage
CREATE INDEX IF NOT EXISTS idx_token_usage_company_id ON token_usage(company_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_conversation_id ON token_usage(conversation_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_model ON token_usage(model);
CREATE INDEX IF NOT EXISTS idx_token_usage_provider ON token_usage(provider);
CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_token_usage_company_created ON token_usage(company_id, created_at DESC);

-- Comments
COMMENT ON TABLE subscription_transactions IS 'Tracks all subscription events (create, upgrade, downgrade, cancel, renew, status changes)';
COMMENT ON TABLE chat_usage_logs IS 'Aggregate daily chat usage statistics per company (no message content stored)';
COMMENT ON TABLE token_usage IS 'Tracks AI token consumption per chat completion for cost analysis and billing';
