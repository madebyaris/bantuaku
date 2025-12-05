#!/bin/bash
# Fix admin user: ensure email_verified=true and company exists

set -e

echo "Fixing admin user..."

# Try docker compose first, fallback to docker-compose
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Update admin user to have email_verified=true and status=active
$DOCKER_COMPOSE exec -T db psql -U bantuaku -d bantuaku_dev <<EOF
-- Update admin user to be verified and active
UPDATE users 
SET email_verified = true, 
    status = 'active',
    role = 'super_admin'
WHERE email = 'admin@bantuaku.id';

-- Ensure company exists for admin user
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

-- Verify the update
SELECT id, email, role, email_verified, status FROM users WHERE email = 'admin@bantuaku.id';
SELECT id, owner_user_id, name, status FROM companies WHERE owner_user_id = 'super-admin-001';
EOF

echo ""
echo "Admin user fixed! Try logging in with:"
echo "  Email: admin@bantuaku.id"
echo "  Password: demo123"
