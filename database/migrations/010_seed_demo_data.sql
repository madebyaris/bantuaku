-- Bantuaku Demo Data Seed
-- This creates a demo company with sample products and sales for hackathon demo
-- Updated to use company_id instead of store_id (after migration 003_add_chat_tables.sql)

-- Demo User (password: demo123)
INSERT INTO users (id, email, password_hash, created_at)
VALUES (
    'demo-user-001',
    'demo@bantuaku.id',
    '$2a$10$E/KmS9sT76xcwUeji.gEDeikxK99miVSTZ9XCLrzcLYayVzvMT1JK',
    NOW()
) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- Demo Company (stores table was renamed to companies in migration 003)
INSERT INTO companies (id, owner_user_id, name, industry, location, subscription_plan, status, created_at)
VALUES (
    'demo-store-001',
    'demo-user-001',
    'Toko Berkah Jaya',
    'retail',
    'Jakarta',
    'pro',
    'active',
    NOW()
) ON CONFLICT DO NOTHING;

-- Sample Products (store_id renamed to company_id, product_name renamed to name in migration 003)
INSERT INTO products (id, company_id, name, sku, category, unit_price, cost, created_at, updated_at)
VALUES
    ('prod-001', 'demo-store-001', 'Kopi Arabica Premium 250g', 'KOP-ARB-250', 'Minuman', 85000, 45000, NOW(), NOW()),
    ('prod-002', 'demo-store-001', 'Teh Hijau Organik 100g', 'TEH-HIJ-100', 'Minuman', 35000, 18000, NOW(), NOW()),
    ('prod-003', 'demo-store-001', 'Gula Aren Bubuk 500g', 'GUL-ARE-500', 'Bahan Makanan', 42000, 22000, NOW(), NOW()),
    ('prod-004', 'demo-store-001', 'Madu Hutan Asli 350ml', 'MAD-HUT-350', 'Makanan Sehat', 125000, 75000, NOW(), NOW()),
    ('prod-005', 'demo-store-001', 'Keripik Pisang Coklat 200g', 'KER-PIS-200', 'Snack', 28000, 14000, NOW(), NOW()),
    ('prod-006', 'demo-store-001', 'Sambal Bawang Premium 150g', 'SAM-BAW-150', 'Bumbu', 32000, 16000, NOW(), NOW()),
    ('prod-007', 'demo-store-001', 'Kacang Mete Panggang 250g', 'KAC-MET-250', 'Snack', 65000, 35000, NOW(), NOW()),
    ('prod-008', 'demo-store-001', 'Minyak Kelapa VCO 500ml', 'MIN-VCO-500', 'Makanan Sehat', 95000, 55000, NOW(), NOW()),
    ('prod-009', 'demo-store-001', 'Rendang Daging Kemasan 250g', 'REN-DAG-250', 'Makanan Siap Saji', 75000, 42000, NOW(), NOW()),
    ('prod-010', 'demo-store-001', 'Abon Sapi Original 150g', 'ABN-SAP-150', 'Makanan Siap Saji', 45000, 25000, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Sample Sales Data (last 60 days)
DO $$
DECLARE
    day_offset INT;
    prod_id VARCHAR(36);
    products_arr VARCHAR[] := ARRAY['prod-001', 'prod-002', 'prod-003', 'prod-004', 'prod-005', 'prod-006', 'prod-007', 'prod-008', 'prod-009', 'prod-010'];
    base_prices NUMERIC[] := ARRAY[85000, 35000, 42000, 125000, 28000, 32000, 65000, 95000, 75000, 45000];
    qty INT;
    sale_price NUMERIC;
BEGIN
    FOR day_offset IN 1..60 LOOP
        FOR i IN 1..array_length(products_arr, 1) LOOP
            -- Random quantity based on product popularity
            qty := FLOOR(RANDOM() * 5 + 1)::INT;
            
            -- Add weekend boost
            IF EXTRACT(DOW FROM (NOW() - (day_offset || ' days')::INTERVAL)) IN (0, 6) THEN
                qty := qty + FLOOR(RANDOM() * 3)::INT;
            END IF;
            
            -- Add month-end boost
            IF EXTRACT(DAY FROM (NOW() - (day_offset || ' days')::INTERVAL)) > 25 THEN
                qty := qty + FLOOR(RANDOM() * 2)::INT;
            END IF;
            
            sale_price := base_prices[i];
            
            -- Insert sale record (store_id renamed to company_id in migration 003)
            INSERT INTO sales_history (company_id, product_id, quantity, price, sale_date, source, created_at)
            VALUES (
                'demo-store-001',
                products_arr[i],
                qty,
                sale_price,
                (NOW() - (day_offset || ' days')::INTERVAL)::DATE,
                CASE 
                    WHEN RANDOM() < 0.6 THEN 'manual'
                    WHEN RANDOM() < 0.8 THEN 'woocommerce'
                    ELSE 'csv'
                END,
                NOW()
            );
        END LOOP;
    END LOOP;
END $$;

-- Sample Integration (WooCommerce connected) - store_id renamed to company_id in migration 003
INSERT INTO integrations (id, company_id, platform, status, last_sync, metadata, created_at)
VALUES (
    'int-001',
    'demo-store-001',
    'woocommerce',
    'connected',
    NOW() - INTERVAL '2 hours',
    '{"store_url": "https://demo-woo.bantuaku.id", "consumer_key": "ck_demo", "consumer_secret": "cs_demo"}',
    NOW()
) ON CONFLICT DO NOTHING;
