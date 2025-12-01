# Remove Inventory Features - Todo List

## Phase 1: Database Migration

- [ ] Create migration file `003_remove_stock.sql` (15min)
  - **Acceptance criteria:** Migration drops `stock` column from products table
  - **Files:** `database/migrations/003_remove_stock.sql`
  - **Dependencies:** None

- [ ] Test migration on dev database (10min)
  - **Acceptance criteria:** Migration runs successfully, stock column removed
  - **Files:** Test migration manually
  - **Dependencies:** Migration file created

## Phase 2: Backend Models

- [ ] Remove Stock field from Product model (5min)
  - **Acceptance criteria:** Product struct has no Stock field
  - **Files:** `backend/models/models.go`
  - **Dependencies:** None

- [ ] Remove CurrentStock from Recommendation model (5min)
  - **Acceptance criteria:** Recommendation struct has no CurrentStock field
  - **Files:** `backend/models/models.go`
  - **Dependencies:** None

- [ ] Remove LowStockCount from DashboardSummary model (5min)
  - **Acceptance criteria:** DashboardSummary struct has no LowStockCount field
  - **Files:** `backend/models/models.go`
  - **Dependencies:** None

## Phase 3: Backend Handlers - Products

- [ ] Remove stock from CreateProductRequest (5min)
  - **Acceptance criteria:** CreateProductRequest has no Stock field
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from UpdateProductRequest (5min)
  - **Acceptance criteria:** UpdateProductRequest has no Stock field
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from ListProducts SQL query (5min)
  - **Acceptance criteria:** SELECT query doesn't include stock column
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from GetProduct SQL query (5min)
  - **Acceptance criteria:** SELECT query doesn't include stock column
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from CreateProduct SQL INSERT (5min)
  - **Acceptance criteria:** INSERT statement doesn't include stock
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from UpdateProduct SQL UPDATE (5min)
  - **Acceptance criteria:** UPDATE statement doesn't include stock
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from Scan operations (5min)
  - **Acceptance criteria:** rows.Scan doesn't include stock variable
  - **Files:** `backend/handlers/products.go`
  - **Dependencies:** Models updated

## Phase 4: Backend Handlers - Forecasts

- [ ] Remove CurrentStock from ForecastResponse (5min)
  - **Acceptance criteria:** ForecastResponse has no CurrentStock field
  - **Files:** `backend/handlers/forecasts.go`
  - **Dependencies:** Models updated

- [ ] Remove stock query from GetForecast (5min)
  - **Acceptance criteria:** No SELECT stock query in GetForecast handler
  - **Files:** `backend/handlers/forecasts.go`
  - **Dependencies:** Models updated

- [ ] Refactor GetRecommendations to demand-only (30min)
  - **Acceptance criteria:** Recommendations show projected demand only, no stock comparison, no "restock" language
  - **Files:** `backend/handlers/forecasts.go`
  - **Dependencies:** Models updated
  - **Notes:** Change from "restock X units" to "projected demand: X units"

## Phase 5: Backend Handlers - Dashboard

- [ ] Remove low stock count query (5min)
  - **Acceptance criteria:** Dashboard query doesn't count low stock products
  - **Files:** `backend/handlers/dashboard.go`
  - **Dependencies:** Models updated

- [ ] Remove LowStockCount from response (5min)
  - **Acceptance criteria:** DashboardSummary response has no low_stock_count
  - **Files:** `backend/handlers/dashboard.go`
  - **Dependencies:** Models updated

## Phase 6: Backend Handlers - Integrations

- [ ] Remove stock sync from WooCommerce import (10min)
  - **Acceptance criteria:** WooCommerce sync doesn't import stock quantity
  - **Files:** `backend/handlers/integrations.go`
  - **Dependencies:** Models updated

## Phase 7: Backend Handlers - AI

- [ ] Remove inventory references from AI prompts (15min)
  - **Acceptance criteria:** AI system prompt doesn't mention inventory/stock
  - **Files:** `backend/handlers/ai.go`
  - **Dependencies:** Models updated

- [ ] Remove stock from StoreContext (10min)
  - **Acceptance criteria:** StoreContext has no stock-related fields
  - **Files:** `backend/handlers/ai.go`
  - **Dependencies:** Models updated

- [ ] Remove low stock items query from AI context (10min)
  - **Acceptance criteria:** AI context doesn't include low stock items
  - **Files:** `backend/handlers/ai.go`
  - **Dependencies:** Models updated

- [ ] Update AI response generation (10min)
  - **Acceptance criteria:** AI responses don't mention stock/inventory
  - **Files:** `backend/handlers/ai.go`
  - **Dependencies:** Models updated

## Phase 8: Frontend Types

- [ ] Remove stock from Product type (5min)
  - **Acceptance criteria:** Product interface has no stock property
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** Backend updated

- [ ] Remove current_stock from Recommendation type (5min)
  - **Acceptance criteria:** Recommendation interface has no current_stock property
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** Backend updated

- [ ] Remove low_stock_count from DashboardSummary type (5min)
  - **Acceptance criteria:** DashboardSummary interface has no low_stock_count property
  - **Files:** `frontend/src/lib/api.ts`
  - **Dependencies:** Backend updated

## Phase 9: Frontend Components - Products Page

- [ ] Remove stock display from product list (10min)
  - **Acceptance criteria:** Product cards don't show stock information
  - **Files:** `frontend/src/pages/ProductsPage.tsx`
  - **Dependencies:** Types updated

- [ ] Remove stock input from product form (10min)
  - **Acceptance criteria:** Add/edit product form has no stock field
  - **Files:** `frontend/src/pages/ProductsPage.tsx`
  - **Dependencies:** Types updated

- [ ] Remove low stock warnings (10min)
  - **Acceptance criteria:** No red badges or warnings for low stock
  - **Files:** `frontend/src/pages/ProductsPage.tsx`
  - **Dependencies:** Types updated

- [ ] Remove stock from product detail view (10min)
  - **Acceptance criteria:** Product detail modal doesn't show stock
  - **Files:** `frontend/src/pages/ProductsPage.tsx`
  - **Dependencies:** Types updated

- [ ] Update forecast display (remove stock comparison) (10min)
  - **Acceptance criteria:** Forecast shows projected demand only, no "need to order X" messages
  - **Files:** `frontend/src/pages/ProductsPage.tsx`
  - **Dependencies:** Types updated

## Phase 10: Frontend Components - Dashboard

- [ ] Remove "Stok Rendah" KPI card (10min)
  - **Acceptance criteria:** Dashboard doesn't show low stock count card
  - **Files:** `frontend/src/pages/DashboardPage.tsx`
  - **Dependencies:** Types updated

- [ ] Remove stock from recommendations display (10min)
  - **Acceptance criteria:** Recommendations panel doesn't show current stock
  - **Files:** `frontend/src/pages/DashboardPage.tsx`
  - **Dependencies:** Types updated

## Phase 11: Documentation

- [ ] Update README.md (15min)
  - **Acceptance criteria:** README doesn't mention inventory management, focuses on forecasting
  - **Files:** `README.md`
  - **Dependencies:** Implementation complete

- [ ] Update feature briefs (10min)
  - **Acceptance criteria:** Feature briefs reflect removal of inventory features
  - **Files:** `specs/active/feat-002-bantuaku-mvp/feature-brief.md`
  - **Dependencies:** Implementation complete

- [ ] Update index.html meta description (5min)
  - **Acceptance criteria:** Meta description doesn't mention inventory
  - **Files:** `frontend/index.html`
  - **Dependencies:** Implementation complete

## Phase 12: Testing & Verification

- [ ] Test product CRUD operations (10min)
  - **Acceptance criteria:** Can create, read, update, delete products without stock
  - **Files:** Manual testing
  - **Dependencies:** All phases complete

- [ ] Test recommendations endpoint (10min)
  - **Acceptance criteria:** Recommendations return demand-only data, no stock fields
  - **Files:** Manual testing
  - **Dependencies:** All phases complete

- [ ] Test dashboard endpoint (10min)
  - **Acceptance criteria:** Dashboard returns no low_stock_count
  - **Files:** Manual testing
  - **Dependencies:** All phases complete

- [ ] Test AI assistant (10min)
  - **Acceptance criteria:** AI doesn't mention inventory/stock in responses
  - **Files:** Manual testing
  - **Dependencies:** All phases complete

- [ ] Verify no stock references in codebase (10min)
  - **Acceptance criteria:** Grep for "stock" shows only comments/docs about removal
  - **Files:** Codebase search
  - **Dependencies:** All phases complete

## Summary

**Total Todos:** 42
**Estimated Time:** ~6 hours
**Status:** Ready to implement
