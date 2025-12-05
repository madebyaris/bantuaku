# Subscription Plans CRUD Feature Brief

## 🎯 Context (2min)
**Problem**: Admin cannot manage subscription plans through the UI. Currently, plans are seeded via SQL and have limited configuration for usage-based limits (chat, file uploads, forecasting).

**Users**: Super Admins managing the platform

**Success**: 
- Full CRUD for subscription plans via admin panel
- Usage-based limits configurable per plan (chat messages, file uploads, forecasting refreshes)
- Plans display clearly with all limits visible

## 🔍 Quick Research (15min)

### Existing Patterns
- `UsersPage.tsx` → CRUD pattern with table, modals, forms | Reuse: ✅
- `backend/handlers/admin/users.go` → Admin handler pattern | Reuse: ✅
- `subscription_plans` table → Already exists, needs new columns | Extend: ✅

### Current Schema (subscription_plans)
```sql
- id, name, display_name
- price_monthly, price_yearly, currency
- max_stores, max_products
- features (JSONB)
- stripe_price_id_monthly/yearly
- is_active, created_at, updated_at
```

### New Fields Needed
| Field | Type | Description |
|-------|------|-------------|
| `max_chats_per_month` | INT | Chat messages limit (NULL = unlimited) |
| `max_file_uploads_per_month` | INT | File upload count limit |
| `max_file_size_mb` | INT | Max file size in MB |
| `max_forecast_refreshes_per_month` | INT | Forecasting refresh limit |

### Tech Decision
**Approach**: Add new columns to existing table, create new admin page (SubscriptionPlansPage)
**Why**: Keeps subscriptions (user records) separate from plans (configuration)
**Avoid**: Modifying SubscriptionsPage (that's for user subscription records)

## ✅ Requirements (10min)

1. **List Plans**: Table showing all plans with name, price, limits, status
2. **Create Plan**: Modal form with all fields including new limits
3. **Edit Plan**: Update existing plan details and limits
4. **Delete/Deactivate**: Soft delete via is_active flag
5. **Display Limits**: Clear visualization of all usage limits

## 🏗️ Implementation (5min)

**Database Migration**: `015_add_subscription_plan_limits.sql`
- Add 4 new columns for usage limits

**Backend Handlers**: `backend/handlers/admin/plans.go`
- ListPlans, GetPlan, CreatePlan, UpdatePlan, DeletePlan

**Frontend Page**: `admin/src/pages/SubscriptionPlansPage.tsx`
- Table with CRUD operations
- Modal for create/edit

**API Client**: Update `admin/src/lib/api.ts`
- Add plans endpoints

**Routes**: Update `admin/src/App.tsx`
- Add /plans route

## 📋 Next Actions (2min)

- [ ] Create database migration for new limit columns (15min)
- [ ] Create backend CRUD handlers for plans (45min)
- [ ] Create frontend SubscriptionPlansPage (1.5h)
- [ ] Add navigation and routes (15min)
- [ ] Test full CRUD flow (15min)

**Start Coding In**: Now

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-01-05
