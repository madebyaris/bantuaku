# feat-002-bantuaku-mvp Feature Brief

## 🎯 Context (2min)
**Problem**: UMKM Indonesia lose IDR 3-5M/month due to poor inventory management - manual tracking, no forecasting, blind to market signals
**Users**: Indonesian UMKM owners (65M+ market), especially those without e-commerce (50%+)
**Success**: MVP demo-ready in 7 days with manual input, WooCommerce sync, 30-day forecasts, AI chat in Bahasa Indonesia

## 🔍 Quick Research (15min)
### Existing Patterns
- Go net/http → Simple, fast, minimal dependencies | Reuse: Standard library approach
- React + Vite + Tailwind → Modern DX, shadcn-style components | Reuse: Component library patterns
- PostgreSQL → Robust, time-series friendly | Reuse: Standard relational patterns
- Vertical Slice Architecture → Business-first organization | Reuse: Feature-based structure

### Tech Decision
**Approach**: Vertical Driven Development (VDD) - build complete slices (UI → Logic → Data) per feature
**Why**: Faster delivery, business-focused, minimal coupling between features
**Avoid**: Layer-first architecture, premature abstractions

## ✅ Requirements (10min)
- **US-001**: User can register/login → JWT auth, store creation, multi-tenant isolation
- **US-002**: User can add products manually → CRUD API + React form, store-scoped
- **US-003**: User can record sales (manual/CSV) → Form + CSV parser, forecast cache invalidation
- **US-004**: User can connect WooCommerce → OAuth-like flow, sync products/orders, status tracking
- **US-005**: System shows 30-day forecasts → Ensemble algorithm (SMA + Exponential Smoothing + Trend), Redis caching
- **US-006**: System provides restock recommendations → Risk-based (high/medium/low), Bahasa Indonesia reasons
- **US-007**: User can ask AI questions → OpenAI integration, Bahasa Indonesia responses, context-aware
- **US-008**: Dashboard shows KPIs → Revenue, products, low stock, forecast accuracy

## 🏗️ Implementation (5min)
**Components**: 
- Backend: 8 handlers (auth, products, sales, integrations, forecasts, sentiment, ai, dashboard)
- Frontend: 6 pages (login, register, dashboard, products, data-input, integrations, ai-chat)
- Database: 11 tables (users, stores, products, sales_history, forecasts, integrations, etc.)

**APIs**: 
- `/api/v1/auth/*` (register, login)
- `/api/v1/products/*` (CRUD)
- `/api/v1/sales/*` (manual, import-csv)
- `/api/v1/integrations/woocommerce/*` (connect, sync-status, sync-now)
- `/api/v1/forecasts/{id}`, `/api/v1/recommendations`
- `/api/v1/ai/analyze`
- `/api/v1/dashboard/summary`

**Data**: PostgreSQL schema with migrations, demo seed data (10 products, 60 days sales)

## 📋 Next Actions (2min)
- [x] Platform foundation (Docker, Go backend, React frontend) ✅
- [x] Auth & store onboarding ✅
- [x] Manual & CSV data input ✅
- [x] WooCommerce integration ✅
- [x] Forecasting & recommendations ✅
- [x] Sentiment & market insights (MVP level) ✅
- [x] AI assistant (Bahasa Indonesia) ✅
- [x] Dashboard & demo narrative ✅

**Start Coding In**: Already completed! 🎉

---
**Total Planning Time**: Retrospective | **Owner**: Development Team | 2025-12-01

## 🔄 Implementation Tracking

### Progress
- [x] All 8 vertical slices implemented
- [x] Backend API complete (20+ endpoints)
- [x] Frontend pages complete (6 pages)
- [x] Database schema & migrations
- [x] Demo data seeded
- [x] Docker setup working

### Blockers
- None - all features delivered

### Notes
- Built using VDD principles - each slice is complete and independent
- Demo account: demo@bantuaku.id / demo123
- All features working and tested manually
