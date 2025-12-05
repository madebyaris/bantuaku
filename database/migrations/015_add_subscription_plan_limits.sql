-- Migration 015: Add usage-based limits to subscription plans
-- Adds columns for chat, file upload, and forecasting limits

-- Add new limit columns
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS max_chats_per_month INT;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS max_file_uploads_per_month INT;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS max_file_size_mb INT DEFAULT 10;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS max_forecast_refreshes_per_month INT;

-- Update existing plans with default limits
UPDATE subscription_plans SET
    max_chats_per_month = 50,
    max_file_uploads_per_month = 5,
    max_file_size_mb = 5,
    max_forecast_refreshes_per_month = 10
WHERE name = 'free';

UPDATE subscription_plans SET
    max_chats_per_month = 500,
    max_file_uploads_per_month = 50,
    max_file_size_mb = 25,
    max_forecast_refreshes_per_month = 100
WHERE name = 'pro';

UPDATE subscription_plans SET
    max_chats_per_month = NULL,  -- Unlimited
    max_file_uploads_per_month = NULL,  -- Unlimited
    max_file_size_mb = 100,
    max_forecast_refreshes_per_month = NULL  -- Unlimited
WHERE name = 'enterprise';

-- Add comments
COMMENT ON COLUMN subscription_plans.max_chats_per_month IS 'Maximum AI chat messages per month (NULL = unlimited)';
COMMENT ON COLUMN subscription_plans.max_file_uploads_per_month IS 'Maximum file uploads per month (NULL = unlimited)';
COMMENT ON COLUMN subscription_plans.max_file_size_mb IS 'Maximum file size in MB';
COMMENT ON COLUMN subscription_plans.max_forecast_refreshes_per_month IS 'Maximum forecast refreshes per month (NULL = unlimited)';
