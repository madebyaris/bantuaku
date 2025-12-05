-- Migration: Add settings table for application-wide configuration
-- Created: 2025-12-05
-- Purpose: Store AI provider preference and other admin-configurable settings

CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on key for fast lookups
CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);

-- Insert default AI provider setting (OpenRouter)
INSERT INTO settings (key, value)
VALUES ('ai_provider', '{"provider": "openrouter"}'::jsonb)
ON CONFLICT (key) DO NOTHING;

-- Add comment
COMMENT ON TABLE settings IS 'Application-wide settings configurable via admin panel';
COMMENT ON COLUMN settings.key IS 'Setting key (e.g., ai_provider)';
COMMENT ON COLUMN settings.value IS 'Setting value as JSON (e.g., {"provider": "openrouter"})';
