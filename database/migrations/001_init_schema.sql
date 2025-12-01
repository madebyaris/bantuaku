-- Bantuaku SaaS - Initial Database Schema
-- PostgreSQL 16

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Stores table
CREATE TABLE IF NOT EXISTS stores (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    location VARCHAR(255),
    subscription_plan VARCHAR(20) DEFAULT 'free',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stores_user_id ON stores(user_id);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_name VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    category VARCHAR(100),
    unit_price NUMERIC(12, 2) DEFAULT 0,
    cost NUMERIC(12, 2) DEFAULT 0,
    stock INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_store_id ON products(store_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_store_sku ON products(store_id, sku) WHERE sku IS NOT NULL AND sku != '';

-- Sales History table
CREATE TABLE IF NOT EXISTS sales_history (
    id BIGSERIAL PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    price NUMERIC(12, 2) DEFAULT 0,
    sale_date DATE NOT NULL,
    source VARCHAR(32) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sales_product_date ON sales_history(product_id, sale_date DESC);
CREATE INDEX IF NOT EXISTS idx_sales_store_date ON sales_history(store_id, sale_date DESC);

-- Forecasts table (cached forecasts)
CREATE TABLE IF NOT EXISTS forecasts (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    forecast_30d INT,
    forecast_60d INT,
    forecast_90d INT,
    confidence REAL,
    eoq NUMERIC(10, 2),
    safety_stock INT,
    algorithm VARCHAR(50),
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_forecasts_product ON forecasts(product_id);
CREATE INDEX IF NOT EXISTS idx_forecasts_expires ON forecasts(expires_at);

-- Recommendations table
CREATE TABLE IF NOT EXISTS recommendations (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    recommended_qty INT,
    reason TEXT,
    risk_level VARCHAR(16),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recommendations_product ON recommendations(product_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_created ON recommendations(created_at);

-- Integrations table (WooCommerce, Shopee, etc.)
CREATE TABLE IF NOT EXISTS integrations (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    status VARCHAR(32) DEFAULT 'disconnected',
    last_sync TIMESTAMPTZ,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_integrations_store_platform ON integrations(store_id, platform);

-- Sentiment Data table
CREATE TABLE IF NOT EXISTS sentiment_data (
    id BIGSERIAL PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id VARCHAR(36) REFERENCES products(id) ON DELETE CASCADE,
    source VARCHAR(50),
    sentiment_score REAL,
    text_snippet TEXT,
    extracted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sentiment_product_date ON sentiment_data(product_id, created_at DESC);

-- Market Trends table
CREATE TABLE IF NOT EXISTS market_trends (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    category VARCHAR(100),
    trend_name VARCHAR(255),
    trend_score REAL,
    source VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_trends_category_date ON market_trends(category, created_at DESC);

-- API Logs table (for monitoring)
CREATE TABLE IF NOT EXISTS api_logs (
    id BIGSERIAL PRIMARY KEY,
    store_id VARCHAR(36),
    endpoint VARCHAR(255),
    status_code INT,
    response_time_ms INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_logs_store_date ON api_logs(store_id, created_at DESC);

-- Documents table (uploaded files for non-website stores)
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(36) PRIMARY KEY,
    store_id VARCHAR(36) NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    period_start DATE,
    period_end DATE,
    status VARCHAR(32) DEFAULT 'uploaded',
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_documents_store_created ON documents(store_id, created_at DESC);
