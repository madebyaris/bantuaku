# Dashboard Improvement Feature Brief

## 🎯 Context (2min)
**Problem**: Dashboard page (`/dashboard`) shows outdated features from old architecture (products, data-input, restocking) and is missing new AI-chat-first features (conversations, insights, company profile).

**Users**: UMKM owners who use Bantuaku platform - need clear overview and quick access to AI chat and insights.

**Success**: Dashboard accurately reflects current AI-chat-first architecture, guides users to primary features (AI chat, insights), and shows relevant company data.

## 🔍 Quick Research (15min)

### Existing Patterns

#### Frontend Patterns
- **DashboardPage.tsx** → Current implementation uses Card components, KPI grid layout, chart with recharts
- **API Client** (`api.ts`) → Uses `api.dashboard.summary()`, `api.recommendations.list()`, `api.market.trends()`
- **Component Library** → Card, CardHeader, CardContent, CardTitle from `@/components/ui/card`
- **Charts** → Uses recharts (LineChart, ResponsiveContainer) for sales visualization
- **Navigation** → Links to `/products`, `/data-input`, `/integrations` (all outdated)

#### Backend Patterns
- **Dashboard Handler** (`handlers/dashboard.go`) → Returns `DashboardSummary` with:
  - TotalProducts (from `products` table)
  - RevenueThisMonth (from `sales_history`)
  - RevenueTrend (calculated)
  - TopSellingProduct
  - ForecastAccuracy (hardcoded 78.5%)
- **API Endpoints Available**:
  - `/api/v1/dashboard/summary` ✅ (exists)
  - `/api/v1/chat/conversations` ✅ (exists)
  - `/api/v1/chat/messages` ✅ (exists)
  - `/api/v1/insights` ✅ (exists)
  - `/api/v1/companies/{id}` ✅ (exists, returns CompanyProfile)
  - `/api/v1/files/upload` ✅ (exists)

#### Data Model
- **Company Model** → Replaces Store, has CompanyProfile aggregate
- **Conversation Model** → Has conversations and messages
- **Insight Model** → Has four types: forecast, market_prediction, marketing_recommendation, gov_regulation
- **FileUpload Model** → Tracks uploaded files with status

### Tech Decision
**Approach**: Refactor existing DashboardPage.tsx to use new API endpoints and remove outdated features. Add new widgets for conversations, insights, and company profile.

**Why**: 
- Reuse existing Card/component patterns
- Leverage existing API endpoints
- Minimal architectural changes
- Quick to implement

**Avoid**: 
- Creating new page structure
- Major refactoring of component library
- New API endpoints (use existing ones)

## ✅ Requirements (10min)

### Remove (Outdated Features)
- [ ] Remove "Total Produk" KPI card (products management removed)
- [ ] Remove "Rekomendasi Restok" card (we don't do stock prediction)
- [ ] Remove "Trend Pasar" card (use Market Prediction page instead)
- [ ] Remove link to `/products` page
- [ ] Remove link to `/data-input` page  
- [ ] Remove "Hubungkan Toko Online" quick action (integrations not primary)
- [ ] Remove `api.recommendations.list()` call (endpoint may not exist)
- [ ] Remove `api.market.trends()` call (endpoint may not exist)

### Update (Terminology & Data)
- [ ] Change "store" → "company" terminology throughout
- [ ] Update dashboard API to use `company_id` instead of `store_id`
- [ ] Update KPIs to reflect company-based model
- [ ] Replace mock chart data with real sales data (if available)
- [ ] Update "Forecast Accuracy" to show actual data or remove if not available

### Add (New Features)
- [ ] **Company Profile Card** - Show company name, industry, location, description
- [ ] **Recent Conversations Widget** - List last 3-5 conversations with preview
- [ ] **Insights Summary** - Show count of each insight type (forecast, market, marketing, regulation)
- [ ] **File Uploads Status** - Show recent uploads and processing status
- [ ] **Quick Actions** - Update to: "Start AI Chat", "Upload File", "View Forecast", "View Market Prediction"
- [ ] **Sales Chart** - Keep but use real data from sales_history (if available)

### Improve (UX Enhancements)
- [ ] Add empty states for when no data exists
- [ ] Add loading states for async data
- [ ] Improve chart to show actual sales data
- [ ] Add links to four outcome pages (Forecast, Market Prediction, Marketing, Regulation)
- [ ] Make AI Chat the primary CTA (prominent button)

## 🏗️ Implementation (5min)

### Components
- **DashboardPage.tsx** → Main component (refactor existing)
- **CompanyProfileCard.tsx** → New widget component
- **RecentConversationsWidget.tsx** → New widget component  
- **InsightsSummaryWidget.tsx** → New widget component
- **FileUploadsWidget.tsx** → New widget component

### APIs
- **Backend**: Update `handlers/dashboard.go` to:
  - Use `company_id` instead of `store_id`
  - Return company profile data
  - Return recent conversations count
  - Return insights summary
  - Return file uploads status
  
- **Frontend**: Update `api.ts` to add:
  - `api.chat.conversations.list()` (if not exists)
  - `api.insights.list()` (if not exists)
  - `api.companies.get(id)` (if not exists)
  - `api.files.list()` (if not exists)

### Data Changes
- Update database queries in `dashboard.go`:
  - `stores` → `companies` table
  - `store_id` → `company_id` columns
  - Add queries for conversations, insights, file_uploads

## 📋 Next Actions (2min)

- [ ] **Phase 1: Backend Updates** (2h)
  - Update dashboard handler to use company_id
  - Add queries for conversations, insights, file uploads
  - Update response model to include new data
  
- [ ] **Phase 2: Frontend Refactor** (3h)
  - Remove outdated components and links
  - Update API client calls
  - Create new widget components
  
- [ ] **Phase 3: Integration** (1h)
  - Connect new widgets to API
  - Update chart with real data
  - Test all features

**Start Coding In**: Immediately after brief approval

---
**Total Planning Time**: ~30min | **Owner**: Development Team | 2025-12-02

<!-- Living Document - Update as you code -->

## 🔄 Implementation Tracking

**CRITICAL**: Follow the todo-list systematically. Mark items as complete, document blockers, update progress.

### Progress
- [ ] Track completed items here
- [ ] Update daily

### Blockers
- [ ] Document any blockers

**See**: [.sdd/IMPLEMENTATION_GUIDE.md](mdc:.sdd/IMPLEMENTATION_GUIDE.md) for detailed execution rules.
