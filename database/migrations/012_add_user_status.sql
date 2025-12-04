-- Bantuaku - Add User Status Field
-- Migration 012: User Suspension and Status Management
-- PostgreSQL 18

-- Add status column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
-- Status values: 'active', 'suspended', 'deleted'

-- Create index for status queries
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Update existing users to active status
UPDATE users SET status = 'active' WHERE status IS NULL;

-- Comments
COMMENT ON COLUMN users.status IS 'User account status: active (default), suspended, deleted';

