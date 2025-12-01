# Remove Inventory Features - Technical Plan

## Overview
Remove all inventory/stock tracking features from Bantuaku, converting it to a pure demand forecasting platform based on sales history only.

## Architecture Changes

### Database Layer
- **Migration**: Create `003_remove_stock.sql` to drop `stock` column from `products` table
- **Impact**: Existing stock data will be lost (acceptable for MVP pivot)
- **Alternative**: Keep column but mark as deprecated (not recommended - clean break)

### Backend Layer
1. **Models** (`backend/models/models.go`)
   - Remove `Stock int` from `Product` struct
   - Remove `CurrentStock int` from `Recommendation` struct
   - Remove `SafetyStock int` from `Forecast` struct (optional - can keep for future)
   - Remove `LowStockCount int` from `DashboardSummary` struct

2. **Handlers**
   - **products.go**: Remove stock from CreateProductRequest, UpdateProductRequest, all SQL queries
   - **forecasts.go**: Remove `CurrentStock` from ForecastResponse, remove stock queries
   - **forecasts.go**: Refactor `GetRecommendations` to show projected demand only (no stock comparison)
   - **dashboard.go**: Remove low stock count query
   - **ai.go**: Remove inventory/stock references from prompts and context

3. **WooCommerce Integration** (`handlers/integrations.go`)
   - Remove stock sync from WooCommerce product import

### Frontend Layer
1. **Components**
   - **ProductsPage.tsx**: Remove stock display, stock input field, low stock warnings
   - **DashboardPage.tsx**: Remove "Stok Rendah" KPI card
   - **api.ts**: Remove stock from Product type, Recommendation type

2. **UI Changes**
   - Remove stock column from product tables
   - Remove stock input from product forms
   - Remove low stock alerts/badges

### Documentation
- Update README.md to remove "inventory" mentions
- Update feature briefs to reflect change
- Update API documentation

## Implementation Strategy

### Phase 1: Database (Foundation)
1. Create migration file
2. Test migration on dev database
3. Document data loss (stock values)

### Phase 2: Backend (Core Logic)
1. Update models
2. Update product handlers
3. Refactor recommendations (demand-only)
4. Update dashboard handler
5. Update AI handler

### Phase 3: Frontend (UI)
1. Update TypeScript types
2. Update product pages
3. Update dashboard
4. Remove stock-related UI elements

### Phase 4: Cleanup & Docs
1. Update README
2. Update feature docs
3. Test end-to-end
4. Verify no broken references

## Risk Assessment
- **Low Risk**: Stock data loss (acceptable for MVP pivot)
- **Medium Risk**: Breaking existing integrations (WooCommerce sync)
- **Mitigation**: Test WooCommerce sync after changes

## Success Criteria
- ✅ No `stock` references in codebase (except comments/docs about removal)
- ✅ Recommendations show projected demand only
- ✅ UI shows no stock fields or warnings
- ✅ AI assistant doesn't mention inventory
- ✅ All tests pass (if tests exist)
- ✅ WooCommerce sync still works (without stock)
