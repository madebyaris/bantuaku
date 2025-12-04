-- Bantuaku - Forecasting and Strategies Tables
-- Migration 007: 12-Month Forecasting System
-- PostgreSQL 18
-- Dependencies: 001_init_schema.sql (products table), 006_trends.sql (trends data)

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
CREATE INDEX IF NOT EXISTS idx_forecast_inputs_created ON forecast_inputs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_product_id ON forecasts_monthly(product_id);
CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_date ON forecasts_monthly(forecast_date DESC);
CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_product_month ON forecasts_monthly(product_id, month);
CREATE INDEX IF NOT EXISTS idx_forecasts_monthly_algorithm ON forecasts_monthly(algorithm);

CREATE INDEX IF NOT EXISTS idx_monthly_strategies_product_id ON monthly_strategies(product_id);
CREATE INDEX IF NOT EXISTS idx_monthly_strategies_forecast_id ON monthly_strategies(forecast_id);
CREATE INDEX IF NOT EXISTS idx_monthly_strategies_month ON monthly_strategies(product_id, month);
CREATE INDEX IF NOT EXISTS idx_monthly_strategies_priority ON monthly_strategies(priority);

-- Comments
COMMENT ON TABLE forecast_inputs IS 'Aggregated inputs used for forecasting (sales history, trends, regulations)';
COMMENT ON TABLE forecasts_monthly IS '12-month forecast predictions per product (month 1-12)';
COMMENT ON TABLE monthly_strategies IS 'Actionable strategies generated per month with structured pricing/inventory/marketing actions';
COMMENT ON COLUMN monthly_strategies.actions IS 'Structured JSON with pricing, inventory, marketing actions';

