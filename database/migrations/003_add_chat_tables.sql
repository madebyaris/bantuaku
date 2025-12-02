-- Bantuaku - Add Chat, Ingestion, and Insight Tables
-- Migration 003: AI-Chat-First Architecture
-- PostgreSQL 18

-- Rename stores to companies and add new fields
ALTER TABLE stores RENAME TO companies;
ALTER TABLE companies RENAME COLUMN store_name TO name;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS business_model VARCHAR(100);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS founded_year INT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS location_region VARCHAR(100);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS country VARCHAR(2) DEFAULT 'ID';
ALTER TABLE companies ADD COLUMN IF NOT EXISTS website VARCHAR(255);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS social_media_handles JSONB;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS marketplaces JSONB;
ALTER TABLE companies RENAME COLUMN user_id TO owner_user_id;

-- Update foreign key references from store_id to company_id
ALTER TABLE products RENAME COLUMN store_id TO company_id;
ALTER TABLE sales_history RENAME COLUMN store_id TO company_id;
ALTER TABLE integrations RENAME COLUMN store_id TO company_id;
ALTER TABLE sentiment_data RENAME COLUMN store_id TO company_id;
ALTER TABLE market_trends RENAME COLUMN store_id TO company_id;
ALTER TABLE api_logs RENAME COLUMN store_id TO company_id;
ALTER TABLE documents RENAME COLUMN store_id TO company_id;

-- Update foreign key constraints
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_store_id_fkey;
ALTER TABLE products ADD CONSTRAINT products_company_id_fkey 
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;

ALTER TABLE sales_history DROP CONSTRAINT IF EXISTS sales_history_store_id_fkey;
ALTER TABLE sales_history ADD CONSTRAINT sales_history_company_id_fkey 
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;

-- Data Sources (Channels/Connectors)
CREATE TABLE IF NOT EXISTS data_sources (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- 'manual', 'csv', 'xlsx', 'pdf', 'marketplace', 'google_trends', 'regulation'
    provider VARCHAR(100),  -- 'tokopedia', 'shopee', 'bukalapak', 'google_trends', 'peraturan_go_id'
    meta JSONB,  -- Account name, URLs, etc.
    status VARCHAR(32) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_data_sources_company_id ON data_sources(company_id);
CREATE INDEX IF NOT EXISTS idx_data_sources_type ON data_sources(type);

-- File Uploads (Ingestion)
CREATE TABLE IF NOT EXISTS file_uploads (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL,  -- 'csv', 'xlsx', 'pdf'
    original_filename VARCHAR(255) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100),
    size_bytes BIGINT NOT NULL,
    status VARCHAR(32) DEFAULT 'uploaded',  -- 'uploaded', 'processing', 'processed', 'failed'
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_file_uploads_company_id ON file_uploads(company_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_status ON file_uploads(status);

-- Conversations (Chat System)
CREATE TABLE IF NOT EXISTS conversations (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    purpose VARCHAR(50),  -- 'onboarding', 'forecasting', 'market_research', 'analysis'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_company_id ON conversations(company_id);
CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);

-- Messages (Chat Messages)
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(36) PRIMARY KEY,
    conversation_id VARCHAR(36) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender VARCHAR(20) NOT NULL,  -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,
    structured_payload JSONB,  -- Extracted fields, tool calls
    file_upload_id VARCHAR(36) REFERENCES file_uploads(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

-- Insights (Four Outcome Types)
CREATE TABLE IF NOT EXISTS insights (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,  -- 'forecast', 'market_prediction', 'marketing_recommendation', 'gov_regulation'
    input_context JSONB,  -- Time ranges, assumptions, filters
    result JSONB,  -- Numbers, charts, recommended actions
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insights_company_id ON insights(company_id);
CREATE INDEX IF NOT EXISTS idx_insights_type ON insights(type);
CREATE INDEX IF NOT EXISTS idx_insights_created_at ON insights(created_at);

-- Update sales_history to reference data_sources and file_uploads
ALTER TABLE sales_history ADD COLUMN IF NOT EXISTS data_source_id VARCHAR(36) REFERENCES data_sources(id);
ALTER TABLE sales_history ADD COLUMN IF NOT EXISTS file_upload_id VARCHAR(36) REFERENCES file_uploads(id);

-- Update products table structure (if needed)
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE products RENAME COLUMN product_name TO name;

-- Update indexes
DROP INDEX IF EXISTS idx_stores_user_id;
CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_companies_country ON companies(country);

DROP INDEX IF EXISTS idx_products_store_id;
CREATE INDEX IF NOT EXISTS idx_products_company_id ON products(company_id);

DROP INDEX IF EXISTS idx_sales_store_date;
CREATE INDEX IF NOT EXISTS idx_sales_company_date ON sales_history(company_id, sale_date DESC);
