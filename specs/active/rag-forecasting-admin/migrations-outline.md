# Database Migrations Outline (004-008)

This document outlines the database migrations required for RAG, Forecasting, Trends, and Admin features.

## Migration 004: pgvector Extension and Embeddings Schema

**File:** `database/migrations/004_pgvector_and_embeddings.sql`

### Purpose
Enable pgvector extension and create base schema for vector embeddings.

### SQL Outline

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table (generic, can be used for any text embeddings)
CREATE TABLE IF NOT EXISTS embeddings (
    id VARCHAR(36) PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,  -- 'regulation_chunk', 'product', etc.
    entity_id VARCHAR(36) NOT NULL,
    embedding vector(1536) NOT NULL,  -- Dimension based on provider (Kolosal.ai = 1536)
    provider VARCHAR(50) NOT NULL DEFAULT 'kolosal',  -- 'kolosal', 'openai', 'cohere'
    model_version VARCHAR(100),  -- Track model version for future migrations
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for KNN search (ivfflat with cosine similarity)
CREATE INDEX IF NOT EXISTS idx_embeddings_vector 
ON embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);  -- Tune based on expected data size (~10K-100K vectors)

-- Index for entity lookups
CREATE INDEX IF NOT EXISTS idx_embeddings_entity ON embeddings(entity_type, entity_id);

-- Comments
COMMENT ON TABLE embeddings IS 'Vector embeddings for semantic search';
COMMENT ON COLUMN embeddings.embedding IS 'Vector embedding (dimension varies by provider)';
COMMENT ON COLUMN embeddings.lists IS 'ivfflat lists parameter - tune based on data size';
```

### Dependencies
- PostgreSQL 11+ (we use 18, satisfied)
- No other migrations required

### Notes
- Vector dimension (1536) matches Kolosal.ai default
- ivfflat lists parameter should be tuned: `rows / 1000` for optimal performance
- Can add more indexes later if needed for different similarity metrics

---

## Migration 005: Regulations Knowledge Base Tables

**File:** `database/migrations/005_regulations_kb.sql`

### Purpose
Create tables for storing regulations data from peraturan.go.id.

### SQL Outline

```sql
-- Regulations table (metadata)
CREATE TABLE IF NOT EXISTS regulations (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    regulation_number VARCHAR(100),  -- e.g., "PP No. 12 Tahun 2023"
    year INT,
    category VARCHAR(100),  -- 'peraturan_pemerintah', 'undang_undang', etc.
    status VARCHAR(50) DEFAULT 'active',  -- 'active', 'revoked', 'amended'
    source_url TEXT NOT NULL,
    pdf_url TEXT,  -- URL to PDF (not stored, only referenced)
    published_date DATE,
    effective_date DATE,
    hash VARCHAR(64),  -- SHA-256 hash for deduplication
    version INT DEFAULT 1,  -- Track updates
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation sources (track where we found it)
CREATE TABLE IF NOT EXISTS regulation_sources (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL DEFAULT 'peraturan_go_id',
    source_url TEXT NOT NULL,
    discovered_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation sections (raw text from PDF)
CREATE TABLE IF NOT EXISTS regulation_sections (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    section_number VARCHAR(50),  -- e.g., "Pasal 1", "Bab II"
    section_title VARCHAR(255),
    content TEXT NOT NULL,  -- Raw extracted text
    page_number INT,
    order_index INT NOT NULL,  -- Order within regulation
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation chunks (semantic chunks for RAG)
CREATE TABLE IF NOT EXISTS regulation_chunks (
    id VARCHAR(36) PRIMARY KEY,
    regulation_id VARCHAR(36) NOT NULL REFERENCES regulations(id) ON DELETE CASCADE,
    section_id VARCHAR(36) REFERENCES regulation_sections(id) ON DELETE CASCADE,
    chunk_text TEXT NOT NULL,
    chunk_index INT NOT NULL,  -- Order within section/regulation
    start_char_offset INT,  -- Character offset in original text
    end_char_offset INT,
    metadata JSONB,  -- Additional metadata (keywords, topics, etc.)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Regulation embeddings (links to embeddings table)
CREATE TABLE IF NOT EXISTS regulation_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    chunk_id VARCHAR(36) NOT NULL REFERENCES regulation_chunks(id) ON DELETE CASCADE,
    embedding_id VARCHAR(36) NOT NULL REFERENCES embeddings(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(chunk_id, embedding_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_regulations_category ON regulations(category);
CREATE INDEX IF NOT EXISTS idx_regulations_year ON regulations(year);
CREATE INDEX IF NOT EXISTS idx_regulations_hash ON regulations(hash);
CREATE INDEX IF NOT EXISTS idx_regulations_status ON regulations(status);

CREATE INDEX IF NOT EXISTS idx_regulation_sources_regulation_id ON regulation_sources(regulation_id);

CREATE INDEX IF NOT EXISTS idx_regulation_sections_regulation_id ON regulation_sections(regulation_id);
CREATE INDEX IF NOT EXISTS idx_regulation_sections_order ON regulation_sections(regulation_id, order_index);

CREATE INDEX IF NOT EXISTS idx_regulation_chunks_regulation_id ON regulation_chunks(regulation_id);
CREATE INDEX IF NOT EXISTS idx_regulation_chunks_section_id ON regulation_chunks(section_id);

CREATE INDEX IF NOT EXISTS idx_regulation_embeddings_chunk_id ON regulation_embeddings(chunk_id);
CREATE INDEX IF NOT EXISTS idx_regulation_embeddings_embedding_id ON regulation_embeddings(embedding_id);

-- Comments
COMMENT ON TABLE regulations IS 'Regulation metadata from peraturan.go.id';
COMMENT ON TABLE regulation_sections IS 'Raw text sections extracted from PDFs';
COMMENT ON TABLE regulation_chunks IS 'Semantic chunks for RAG retrieval';
COMMENT ON TABLE regulation_embeddings IS 'Links regulation chunks to vector embeddings';
```

### Dependencies
- Migration 004 (embeddings table)

### Notes
- PDFs are NOT stored, only extracted text
- Hash field enables deduplication
- Version field tracks regulation updates
- Chunks link to both regulations and sections for flexible retrieval

---

## Migration 006: Google Trends Storage Tables

**File:** `database/migrations/006_trends.sql`

### Purpose
Create tables for storing Google Trends data.

### SQL Outline

```sql
-- Trends keywords (tracked keywords per company)
CREATE TABLE IF NOT EXISTS trends_keywords (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    keyword VARCHAR(255) NOT NULL,
    geo VARCHAR(10) DEFAULT 'ID',  -- Country code (ID = Indonesia)
    category VARCHAR(100),  -- Optional category grouping
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, keyword, geo)
);

-- Trends time series (interest over time)
CREATE TABLE IF NOT EXISTS trends_series (
    id VARCHAR(36) PRIMARY KEY,
    keyword_id VARCHAR(36) NOT NULL REFERENCES trends_keywords(id) ON DELETE CASCADE,
    time TIMESTAMPTZ NOT NULL,  -- Timestamp for data point
    value INT NOT NULL,  -- Interest value (0-100)
    geo VARCHAR(10) DEFAULT 'ID',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(keyword_id, time, geo)
);

-- Trends related queries (related searches)
CREATE TABLE IF NOT EXISTS trends_related_queries (
    id VARCHAR(36) PRIMARY KEY,
    keyword_id VARCHAR(36) NOT NULL REFERENCES trends_keywords(id) ON DELETE CASCADE,
    related_keyword VARCHAR(255) NOT NULL,
    relationship_type VARCHAR(50),  -- 'rising', 'top', 'related'
    value INT,  -- Interest value or growth percentage
    geo VARCHAR(10) DEFAULT 'ID',
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_trends_keywords_company_id ON trends_keywords(company_id);
CREATE INDEX IF NOT EXISTS idx_trends_keywords_active ON trends_keywords(company_id, is_active);

CREATE INDEX IF NOT EXISTS idx_trends_series_keyword_time ON trends_series(keyword_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_trends_series_time ON trends_series(time DESC);

CREATE INDEX IF NOT EXISTS idx_trends_related_keyword_id ON trends_related_queries(keyword_id);
CREATE INDEX IF NOT EXISTS idx_trends_related_captured ON trends_related_queries(keyword_id, captured_at DESC);

-- Comments
COMMENT ON TABLE trends_keywords IS 'Keywords tracked per company for Google Trends';
COMMENT ON TABLE trends_series IS 'Time series interest data from Google Trends';
COMMENT ON TABLE trends_related_queries IS 'Related queries and trending searches';
```

### Dependencies
- Migration 003 (companies table exists)

### Notes
- Time series data stored at daily granularity
- Geo field supports multi-country analysis (default: Indonesia)
- Related queries captured periodically (not every scrape)
- Unique constraints prevent duplicate data

---

## Migration 007: Forecasting and Strategies Tables

**File:** `database/migrations/007_forecasts_strategies.sql`

### Purpose
Create tables for 12-month forecasting and monthly strategies.

### SQL Outline

```sql
-- Forecast inputs (aggregated inputs for forecasting)
CREATE TABLE IF NOT EXISTS forecast_inputs (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    forecast_period_start DATE NOT NULL,
    forecast_period_end DATE NOT NULL,
    sales_history JSONB NOT NULL,  -- Historical sales data
    trends_data JSONB,  -- Google Trends signals
    regulation_flags JSONB,  -- Relevant regulation flags
    exogenous_factors JSONB,  -- Other factors (seasonality, events)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Monthly forecasts (12-month horizon)
CREATE TABLE IF NOT EXISTS forecasts_monthly (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    forecast_input_id VARCHAR(36) REFERENCES forecast_inputs(id),
    month INT NOT NULL CHECK (month >= 1 AND month <= 12),  -- 1-12 months ahead
    forecast_date DATE NOT NULL,  -- Date this forecast was generated
    predicted_quantity INT NOT NULL,
    confidence_lower INT,  -- Lower bound
    confidence_upper INT,  -- Upper bound
    confidence_score REAL,  -- 0-1 confidence
    algorithm VARCHAR(100),  -- 'prophet', 'arima', 'lstm', etc.
    model_version VARCHAR(100),
    metadata JSONB,  -- Additional forecast details
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(product_id, forecast_date, month)
);

-- Monthly strategies (actions per month)
CREATE TABLE IF NOT EXISTS monthly_strategies (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    forecast_id VARCHAR(36) REFERENCES forecasts_monthly(id),
    month INT NOT NULL CHECK (month >= 1 AND month <= 12),
    strategy_text TEXT NOT NULL,  -- Human-readable strategy reasoning
    actions JSONB NOT NULL,  -- Structured actions:
    -- {
    --   "pricing": {"action": "increase", "percentage": 5, "reason": "..."},
    --   "inventory": {"action": "restock", "quantity": 100, "reason": "..."},
    --   "marketing": {"channels": ["social", "email"], "budget": 500000, "reason": "..."}
    -- }
    priority VARCHAR(20) DEFAULT 'medium',  -- 'low', 'medium', 'high', 'critical'
    estimated_impact JSONB,  -- Expected impact metrics
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_forecast_inputs_product_id ON forecast_inputs(product_id);
CREATE INDEX IF NOT EXISTS idx_forecast_inputs_period ON forecast_inputs(forecast_period_start, forecast_period_end);

CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_product_id ON forecasts_monthly(product_id);
CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_date ON forecasts_monthly(forecast_date DESC);
CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_product_month ON forecasts_monthly(product_id, month);

CREATE INDEX IF NOT EXISTS idx_monthly_strategies_product_id ON monthly_strategies(product_id);
CREATE INDEX IF NOT EXISTS idx_monthly_strategies_forecast_id ON monthly_strategies(forecast_id);
CREATE INDEX IF NOT EXISTS idx_monthly_strategies_month ON monthly_strategies(product_id, month);

-- Comments
COMMENT ON TABLE forecast_inputs IS 'Aggregated inputs used for forecasting';
COMMENT ON TABLE forecasts_monthly IS '12-month forecast predictions per product';
COMMENT ON TABLE monthly_strategies IS 'Actionable strategies generated per month';
COMMENT ON COLUMN monthly_strategies.actions IS 'Structured JSON with pricing, inventory, marketing actions';
```

### Dependencies
- Migration 001 (products table exists)
- Migration 006 (trends tables for exogenous signals)

### Notes
- Forecasts stored per month (1-12) for 12-month horizon
- Strategies link to forecasts but can be regenerated independently
- Actions stored as structured JSON for easy parsing
- Confidence intervals support uncertainty visualization

---

## Migration 008: Admin, Billing, and RBAC Tables

**File:** `database/migrations/008_admin_billing_rbac.sql`

### Purpose
Create tables for admin panel, billing, and role-based access control.

### SQL Outline

```sql
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

CREATE INDEX IF NOT EXISTS idx_stripe_webhooks_event_id ON stripe_webhooks(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_stripe_webhooks_processed ON stripe_webhooks(processed, created_at);

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
COMMENT ON TABLE subscription_plans IS 'Available subscription plans';
COMMENT ON TABLE subscriptions IS 'Active subscriptions per company';
COMMENT ON TABLE payments IS 'Payment transaction history';
COMMENT ON TABLE stripe_webhooks IS 'Stripe webhook events for idempotency';
COMMENT ON TABLE audit_logs IS 'Security audit trail for sensitive actions';
```

### Dependencies
- Migration 001 (users table)
- Migration 003 (companies table)

### Notes
- Role added to existing users table (default: 'user')
- Stripe integration ready (test mode initially)
- Audit logs capture all sensitive actions
- Webhook idempotency prevents duplicate processing
- Default plans seeded for immediate use

---

## Migration Execution Order

1. **004** - pgvector_and_embeddings.sql (foundation)
2. **005** - regulations_kb.sql (depends on 004)
3. **006** - trends.sql (independent)
4. **007** - forecasts_strategies.sql (depends on 001, 006)
5. **008** - admin_billing_rbac.sql (depends on 001, 003)

## Rollback Considerations

Each migration should include rollback statements (optional, for development):

```sql
-- Example rollback for 004
-- DROP INDEX IF EXISTS idx_embeddings_vector;
-- DROP TABLE IF EXISTS embeddings;
-- DROP EXTENSION IF EXISTS vector;
```

## Testing Checklist

- [ ] All migrations apply cleanly
- [ ] Indexes created successfully
- [ ] Foreign keys work correctly
- [ ] Unique constraints prevent duplicates
- [ ] pgvector extension functions correctly
- [ ] Vector index performs KNN queries efficiently
- [ ] Rollback scripts work (if implemented)

