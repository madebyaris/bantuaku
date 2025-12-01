-- Migration: Remove stock column from products table
-- Bantuaku is pivoting to demand forecasting only, no inventory management

ALTER TABLE products DROP COLUMN IF EXISTS stock;

-- Note: This will permanently remove stock data. This is intentional for the MVP pivot.
