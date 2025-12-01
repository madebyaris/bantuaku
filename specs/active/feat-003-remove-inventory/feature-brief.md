# feat-003-remove-inventory Feature Brief

## 🎯 Context (2min)
**Problem**: Bantuaku should focus on demand forecasting only, not inventory management. Stock tracking adds complexity and is out of scope for the core value proposition.
**Users**: Same UMKM users, but platform becomes simpler without inventory tracking
**Success**: All stock/inventory references removed, forecasting works purely on sales history, UI simplified

## 🔍 Quick Research (15min)
### Existing Patterns
- Database migrations → Use ALTER TABLE to remove/deprecate stock column | Reuse: Migration pattern from 001_init_schema.sql
- Model updates → Remove Stock field from Product struct | Reuse: models/models.go pattern
- Handler refactoring → Remove stock logic from recommendations | Reuse: handlers/forecasts.go pattern
- UI cleanup → Remove stock displays from React components | Reuse: Component pattern from ProductsPage.tsx

### Tech Decision
**Approach**: Systematic removal - database → models → handlers → frontend → AI prompts
**Why**: Ensures no broken references, maintains data integrity, clean separation
**Avoid**: Leaving deprecated fields, breaking existing data, incomplete removal

## ✅ Requirements (10min)
- **REQ-001**: Remove stock field from products table → Migration to drop column or set to NULL
- **REQ-002**: Remove Stock from Product model → Update models/models.go
- **REQ-003**: Remove stock from product CRUD handlers → Update handlers/products.go (create, update, list)
- **REQ-004**: Convert recommendations to demand-only → Remove stock comparison, show projected demand only
- **REQ-005**: Remove low stock tracking from dashboard → Update handlers/dashboard.go, remove low_stock_count
- **REQ-006**: Remove stock UI elements → Update ProductsPage.tsx, DashboardPage.tsx
- **REQ-007**: Update AI assistant prompts → Remove inventory/stock references from handlers/ai.go
- **REQ-008**: Update documentation → README.md, remove inventory mentions

## 🏗️ Implementation (5min)
**Components**: 
- Database: Migration to remove stock column
- Backend: models.go, handlers/products.go, handlers/forecasts.go, handlers/dashboard.go, handlers/ai.go
- Frontend: ProductsPage.tsx, DashboardPage.tsx, api.ts types
- Docs: README.md, specs files

**APIs**: 
- No new APIs, but responses change (remove stock fields)
- `/api/v1/products/*` - Remove stock from request/response
- `/api/v1/recommendations` - Convert to demand-only (remove current_stock)
- `/api/v1/dashboard/summary` - Remove low_stock_count

**Data**: Migration removes stock column, existing data preserved (other fields intact)

## 📋 Next Actions (2min)
- [ ] Create database migration to remove stock column (30min)
- [ ] Update backend models and handlers (1h)
- [ ] Update frontend components (1h)
- [ ] Update AI prompts (30min)
- [ ] Update documentation (30min)

**Start Coding In**: Ready to implement

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-01

## 🔄 Implementation Tracking

### Progress
- [ ] Database migration created
- [ ] Backend models updated
- [ ] Backend handlers updated
- [ ] Frontend components updated
- [ ] AI prompts updated
- [ ] Documentation updated

### Blockers
- None identified
