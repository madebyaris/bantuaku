-- Bantuaku - Google Trends Storage Tables
-- Migration 006: Trends Data Storage
-- PostgreSQL 18
-- Dependencies: 003_add_chat_tables.sql (companies table)

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
CREATE INDEX IF NOT EXISTS idx_trends_keywords_geo ON trends_keywords(geo);

CREATE INDEX IF NOT EXISTS idx_trends_series_keyword_time ON trends_series(keyword_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_trends_series_time ON trends_series(time DESC);
CREATE INDEX IF NOT EXISTS idx_trends_series_geo ON trends_series(geo);

CREATE INDEX IF NOT EXISTS idx_trends_related_keyword_id ON trends_related_queries(keyword_id);
CREATE INDEX IF NOT EXISTS idx_trends_related_captured ON trends_related_queries(keyword_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_trends_related_type ON trends_related_queries(relationship_type);

-- Comments
COMMENT ON TABLE trends_keywords IS 'Keywords tracked per company for Google Trends';
COMMENT ON TABLE trends_series IS 'Time series interest data from Google Trends (daily granularity)';
COMMENT ON TABLE trends_related_queries IS 'Related queries and trending searches captured periodically';

