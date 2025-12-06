-- Bantuaku - Super Admin Seed
-- Migration 011: Create super admin account for admin panel access
-- Password: demo123 (same as demo user for consistency)

-- Super Admin User
INSERT INTO users (id, email, password_hash, role, email_verified, status, created_at)
VALUES (
    'super-admin-001',
    'admin@bantuaku.id',
    '$2a$10$E/KmS9sT76xcwUeji.gEDeikxK99miVSTZ9XCLrzcLYayVzvMT1JK', -- demo123
    'super_admin',
    true, -- Email verified for admin
    'active', -- Active status
    NOW()
) ON CONFLICT (email) DO UPDATE SET 
    password_hash = EXCLUDED.password_hash,
    role = 'super_admin',
    email_verified = true,
    status = 'active';

-- Create a demo company for the super admin (optional, for testing)
INSERT INTO companies (id, owner_user_id, name, industry, subscription_plan, status, created_at)
VALUES (
    'admin-company-001',
    'super-admin-001',
    'Bantuaku Admin Company',
    'technology',
    'enterprise',
    'active',
    NOW()
) ON CONFLICT DO NOTHING;

-- Comments
COMMENT ON COLUMN users.role IS 'User role: user (default), admin, super_admin';

