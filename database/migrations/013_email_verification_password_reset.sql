-- Bantuaku - Email Verification and Password Reset
-- Migration 013: Email Verification with OTP and Password Reset
-- PostgreSQL 18
-- Dependencies: 001_init_schema.sql (users table)

-- Add email verification columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

-- Create verification_codes table for OTPs and password reset tokens
CREATE TABLE IF NOT EXISTS verification_codes (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(5),  -- 5-digit OTP for email verification (NULL for password reset)
    token VARCHAR(64),  -- Secure token for password reset (NULL for OTP)
    code_type VARCHAR(20) NOT NULL,  -- 'email_verification' or 'password_reset'
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_code_or_token CHECK (
        (code IS NOT NULL AND token IS NULL AND code_type = 'email_verification') OR
        (code IS NULL AND token IS NOT NULL AND code_type = 'password_reset')
    )
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_verification_codes_code ON verification_codes(code) WHERE code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_verification_codes_token ON verification_codes(token) WHERE token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_verification_codes_user_id ON verification_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_verification_codes_user_type_expires ON verification_codes(user_id, code_type, expires_at);

-- Create unique constraint to prevent multiple active codes/tokens per user per type
-- Only one active (unused, not expired) code/token per user per type
CREATE UNIQUE INDEX IF NOT EXISTS idx_verification_codes_active_unique 
ON verification_codes(user_id, code_type) 
WHERE used_at IS NULL AND expires_at > NOW();

-- Create index for email_verified queries
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);

-- Comments
COMMENT ON COLUMN users.email_verified IS 'Whether the user email has been verified';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified';
COMMENT ON TABLE verification_codes IS 'Stores OTP codes for email verification and tokens for password reset';
COMMENT ON COLUMN verification_codes.code IS '5-digit OTP for email verification (NULL for password reset)';
COMMENT ON COLUMN verification_codes.token IS 'Secure token for password reset (NULL for OTP)';
COMMENT ON COLUMN verification_codes.code_type IS 'Type: email_verification or password_reset';

