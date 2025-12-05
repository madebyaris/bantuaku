-- Seed super admin user
INSERT INTO users (id, email, password_hash, role, email_verified, status, created_at) 
VALUES (
    'super-admin-001',
    'admin@bantuaku.id',
    '$2a$10$E/KmS9sT76xcwUeji.gEDeikxK99miVSTZ9XCLrzcLYayVzvMT1JK',
    'super_admin',
    true,
    'active',
    NOW()
) ON CONFLICT (email) DO UPDATE SET 
    password_hash = EXCLUDED.password_hash, 
    role = 'super_admin', 
    email_verified = true, 
    status = 'active';

-- Create admin company
INSERT INTO companies (id, owner_user_id, name, industry, subscription_plan, status, created_at) 
VALUES (
    'admin-company-001',
    'super-admin-001',
    'Bantuaku Admin Company',
    'technology',
    'enterprise',
    'active',
    NOW()
) ON CONFLICT (id) DO UPDATE SET 
    status = 'active', 
    owner_user_id = 'super-admin-001';
