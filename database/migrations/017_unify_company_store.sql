-- 017_unify_company_store.sql
-- Ensure all legacy store_id/product_name columns are migrated to company_id/name
-- Safe to re-run: guards each rename with an existence check.

DO $$
BEGIN
  -- Products: store_id -> company_id, product_name -> name
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'products' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE products RENAME COLUMN store_id TO company_id';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'products' AND column_name = 'product_name'
  ) THEN
    EXECUTE 'ALTER TABLE products RENAME COLUMN product_name TO name';
  END IF;

  -- Sales history: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_history' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE sales_history RENAME COLUMN store_id TO company_id';
  END IF;

  -- Integrations: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'integrations' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE integrations RENAME COLUMN store_id TO company_id';
  END IF;

  -- Sentiment data: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sentiment_data' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE sentiment_data RENAME COLUMN store_id TO company_id';
  END IF;

  -- Market trends: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'market_trends' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE market_trends RENAME COLUMN store_id TO company_id';
  END IF;

  -- API logs: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'api_logs' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE api_logs RENAME COLUMN store_id TO company_id';
  END IF;

  -- Documents: store_id -> company_id
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'documents' AND column_name = 'store_id'
  ) THEN
    EXECUTE 'ALTER TABLE documents RENAME COLUMN store_id TO company_id';
  END IF;
END
$$;

-- Ensure indexes exist on company_id
CREATE INDEX IF NOT EXISTS idx_products_company_id ON products(company_id);
CREATE INDEX IF NOT EXISTS idx_sales_company_id ON sales_history(company_id);
CREATE INDEX IF NOT EXISTS idx_integrations_company_id ON integrations(company_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_company_id ON sentiment_data(company_id);
CREATE INDEX IF NOT EXISTS idx_market_trends_company_id ON market_trends(company_id);
CREATE INDEX IF NOT EXISTS idx_api_logs_company_id ON api_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_documents_company_id ON documents(company_id);
