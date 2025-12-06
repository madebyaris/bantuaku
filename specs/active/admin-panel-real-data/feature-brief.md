# Admin Panel Real Data Integration Feature Brief

## 🎯 Context (2min)
**Problem**: Admin panel pages are partially connected to real data but missing key information:
- Dashboard shows stats but may need verification
- Users page doesn't display subscription plan information
- Subscriptions page may not show all relevant details
- Audit logs page needs proper filtering and display
- Settings page needs to ensure all settings are properly loaded and saved

**Users**: 
- **Super Admins**: Need complete visibility into all platform data
- **Admins**: Need accurate data for user and subscription management

**Success**: 
- All admin pages display complete, real-time data from database
- Users page shows subscription plan for each user
- Dashboard stats are accurate and reflect current state
- Subscriptions page shows all subscription details correctly
- Audit logs are filterable and display all relevant information
- Settings page properly loads and saves all configuration

## 🔍 Quick Research (15min)

### Existing Patterns

**1. Admin API Handlers** (`backend/handlers/admin/`)
- **Usage**: RESTful endpoints returning JSON with pagination
- **Pattern**: Handler struct with db, logger, jwtSecret, auditLogger
- **Response Format**: `{ data: [...], pagination: { page, limit, total } }`
- **Reusable**: All admin endpoints follow same pattern

**2. Database Schema Relationships**
- **Users → Companies**: `companies.owner_user_id = users.id`
- **Companies → Subscriptions**: `subscriptions.company_id = companies.id`
- **Companies**: Has `subscription_plan` field (free, pro, enterprise)
- **Subscriptions**: Detailed subscription records with periods, status, Stripe IDs

**3. Frontend API Client** (`admin/src/lib/api.ts`)
- **Usage**: Centralized API client with typed responses
- **Pattern**: `api.admin.{resource}.{action}()` methods
- **Reusable**: All pages use same API client pattern

**4. React Page Components** (`admin/src/pages/`)
- **Usage**: React functional components with hooks
- **Pattern**: `useState` for data/loading, `useEffect` for data fetching
- **Components**: shadcn/ui (Card, Button, Table, Select, Input)
- **Styling**: Tailwind CSS with "Neon Finance" theme

**5. Current Data Flow**
- **Dashboard**: ✅ Uses `api.admin.stats.get()` - returns `{ total_users, total_subscriptions, active_subscriptions, total_audit_logs }`
- **Users**: ✅ Uses `api.admin.users.list()` - returns users with `store_name`, `industry` but **missing subscription_plan**
- **Subscriptions**: ✅ Uses `api.admin.subscriptions.list()` - returns subscriptions with company/plan info
- **Audit Logs**: ✅ Uses `api.admin.auditLogs.list()` - returns logs with filtering
- **Settings**: ✅ Uses `api.admin.settings.getAIProvider()` - returns current AI provider

### Tech Decision
**Approach**: Enhance existing API responses and frontend components to display complete data
**Why**: 
- APIs already exist and work, just need to add missing fields
- Frontend components are structured correctly, need minor updates
- Database schema supports all required relationships
**Avoid**: 
- Creating new endpoints when existing ones can be enhanced
- Mock data or hardcoded values
- Breaking existing functionality

## ✅ Requirements (10min)

**1. Dashboard Page** (`/dashboard`)
- ✅ Display real-time stats from database
- ✅ Show accurate counts for users, subscriptions, audit logs
- ✅ Handle loading and error states gracefully

**2. Users Page** (`/users`)
- ✅ Display user list with all current fields (email, role, status, store_name, industry)
- ✅ **Add subscription plan display** - Show subscription plan from user's company
- ✅ Display subscription status (active, canceled, etc.) if available
- ✅ Show subscription period dates if subscription exists

**3. Subscriptions Page** (`/subscriptions`)
- ✅ Display all subscription records with company and plan information
- ✅ Show subscription status, period dates, Stripe IDs
- ✅ Handle empty state when no subscriptions exist
- ✅ Proper pagination for large subscription lists

**4. Audit Logs Page** (`/audit-logs`)
- ✅ Display audit log entries with all metadata
- ✅ Filter by action, resource_type, user_id
- ✅ Show formatted timestamps and metadata
- ✅ Handle large log volumes with pagination

**5. Settings Page** (`/settings`)
- ✅ Load current AI provider setting from database
- ✅ Save AI provider changes and persist to database
- ✅ Show success/error feedback for save operations
- ✅ Handle loading states during save

## 🏗️ Implementation (5min)

**Components to Update**:
1. `admin/src/pages/DashboardPage.tsx` - Verify data display (likely already correct)
2. `admin/src/pages/UsersPage.tsx` - Add subscription plan column and data
3. `admin/src/pages/SubscriptionsPage.tsx` - Verify all fields display correctly
4. `admin/src/pages/AuditLogsPage.tsx` - Verify filtering and display
5. `admin/src/pages/SettingsPage.tsx` - Verify save/load functionality

**APIs to Enhance**:
1. `backend/handlers/admin/users.go` - `ListUsers()` - Add subscription_plan and subscription status to response
2. `backend/handlers/admin/users.go` - `GetUser()` - Add subscription info to single user response
3. Verify other endpoints return complete data

**Data Changes**:
- Update `User` struct in backend to include subscription fields
- Update frontend `User` interface to include subscription fields
- Update SQL queries to JOIN with subscriptions table

**Database Queries**:
- Users query: JOIN `companies` and `subscriptions` to get plan info
- Subscriptions query: Already joins companies and plans correctly
- Stats query: Already counts correctly from tables

## 📋 Next Actions (2min)

- [ ] Update backend `ListUsers()` handler to include subscription_plan from companies table (15min)
- [ ] Update backend `GetUser()` handler to include subscription details (10min)
- [ ] Update frontend `User` interface to include subscription fields (5min)
- [ ] Update `UsersPage.tsx` to display subscription plan column (20min)
- [ ] Test Users page with real data and verify subscription display (10min)
- [ ] Verify Dashboard stats are accurate (5min)
- [ ] Verify Subscriptions page displays all fields correctly (10min)
- [ ] Verify Audit Logs filtering works correctly (10min)
- [ ] Verify Settings page save/load works correctly (10min)
- [ ] Test all pages end-to-end with real database data (15min)

**Start Coding In**: ~30min

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-01-05

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Track completed items here
- [ ] Update daily

### Blockers
- [ ] Document any blockers

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
