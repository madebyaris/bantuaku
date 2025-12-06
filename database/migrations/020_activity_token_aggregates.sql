-- Migration 020: Activity & Token Usage Aggregates + user_id on token_usage
-- Provides aggregate tables for fast dashboards and adds user_id to token_usage for per-user metrics.

-- Add user_id to token_usage for per-user breakdown (nullable for legacy rows)
ALTER TABLE token_usage
    ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) REFERENCES users(id);

-- Activity aggregates (daily grain)
CREATE TABLE IF NOT EXISTS activity_aggregates (
    date DATE NOT NULL,
    user_id VARCHAR(36),
    company_id VARCHAR(36),
    action_type VARCHAR(100) NOT NULL,
    count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (date, user_id, company_id, action_type)
);

-- Token usage aggregates (daily grain; weekly/monthly can be derived)
CREATE TABLE IF NOT EXISTS token_usage_aggregates (
    date DATE NOT NULL,
    user_id VARCHAR(36),
    company_id VARCHAR(36),
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (date, user_id, company_id, model, provider)
);

-- Indexes for fast filters
CREATE INDEX IF NOT EXISTS idx_activity_agg_date ON activity_aggregates(date DESC);
CREATE INDEX IF NOT EXISTS idx_activity_agg_company ON activity_aggregates(company_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_activity_agg_user ON activity_aggregates(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_activity_agg_action ON activity_aggregates(action_type, date DESC);

CREATE INDEX IF NOT EXISTS idx_token_usage_agg_date ON token_usage_aggregates(date DESC);
CREATE INDEX IF NOT EXISTS idx_token_usage_agg_company ON token_usage_aggregates(company_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_token_usage_agg_user ON token_usage_aggregates(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_token_usage_agg_model ON token_usage_aggregates(model, provider, date DESC);

-- Comments
COMMENT ON TABLE activity_aggregates IS 'Daily aggregates of activity events (chat, upload, RAG, token) grouped by user/company/action';
COMMENT ON TABLE token_usage_aggregates IS 'Daily aggregates of token usage grouped by user/company/model/provider';

