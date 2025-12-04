# RAG, Regulations Scraper, Google Trends, 12‑Month Forecasting, Admin Panel — Todo List

## Phase 0: Discovery & Architecture (PLANNING)

- [ ] Finalize architecture for vector DB, scraping, forecasting, and admin app (4h)
  - **Acceptance criteria:** Clear component diagram, data flow, and tech choices (pgvector vs external vector DB; Go vs Node/Python for scraping/forecasting)
  - **Files:** specs/active/rag-forecasting-admin/ (architecture notes)
  - **Dependencies:** None

## Phase 1: Data Model & Migrations (DB FOUNDATION)

- [ ] Add pgvector extension and embedding schema (3h)
  - **Acceptance criteria:** Database has pgvector installed; embeddings table supports cosine/ivfflat index
  - **Files:** database/migrations/004_pgvector_and_embeddings.sql
  - **Dependencies:** PostgreSQL ready

- [ ] Create regulations knowledge base tables (4h)
  - **Acceptance criteria:** Tables: regulations, regulation_sections, regulation_chunks, regulation_sources, regulation_embeddings
  - **Files:** database/migrations/005_regulations_kb.sql
  - **Dependencies:** 004_pgvector_and_embeddings.sql

- [ ] Create Google Trends storage tables (3h)
  - **Acceptance criteria:** Tables: trends_keywords, trends_series (time, value, geo, keyword_id), trends_related_queries
  - **Files:** database/migrations/006_trends.sql
  - **Dependencies:** None

- [ ] Create forecasting and strategies tables (3h)
  - **Acceptance criteria:** Tables: forecasts_monthly (12‑step horizon), forecast_inputs, monthly_strategies (reasoning + structured actions)
  - **Files:** database/migrations/007_forecasts_strategies.sql
  - **Dependencies:** Sales history exists

- [ ] Create billing/subscription/admin RBAC tables (4h)
  - **Acceptance criteria:** Columns/ tables for user roles; subscription_plans, subscriptions, payments, stripe_webhooks; audit_logs
  - **Files:** database/migrations/008_admin_billing_rbac.sql
  - **Dependencies:** Users table

## Phase 2: Regulations Scraper (peraturan.go.id) — Ingestion

- [x] Build crawler to enumerate regulation PDFs (5h)
  - **Acceptance criteria:** Crawl list/detail pages, discover PDF links with metadata (title, number, year, category, URL)
  - **Files:** backend/services/scraper/regulations/crawler.go (or node script alternative), backend/config/config.go
  - **Dependencies:** None
  - **Status:** ✅ Completed - crawler.go implemented with goquery

- [x] PDF text extraction without persisting PDFs (6h)
  - **Acceptance criteria:** Download to temp, extract clean text (ID locale), discard PDF; store raw text + metadata; handle scanned PDFs with OCR fallback if needed
  - **Files:** backend/services/scraper/regulations/extract.go
  - **Dependencies:** Crawler
  - **Status:** ✅ Completed - extract.go implemented with OCR fallback support

- [x] Chunking and normalization pipeline (3h)
  - **Acceptance criteria:** Split into semantic chunks with overlap; persist to regulation_chunks; link to regulation and source
  - **Files:** backend/services/scraper/regulations/chunker.go
  - **Dependencies:** Extractor
  - **Status:** ✅ Completed - chunker.go implemented with semantic chunking

- [x] Dedup/versioning and idempotent upserts (2h)
  - **Acceptance criteria:** Hashing to avoid duplicates; version field per regulation; safe re‑runs
  - **Files:** backend/services/scraper/regulations/store.go
  - **Dependencies:** Chunker
  - **Status:** ✅ Completed - store.go implemented with SHA-256 hashing

- [x] Scheduler + admin trigger endpoints (3h)
  - **Acceptance criteria:** Daily job; manual run endpoint; progress logs
  - **Files:** backend/handlers/regulations.go, backend/main.go (routes), backend/services/scheduler/scheduler.go
  - **Dependencies:** Ingestion pipeline
  - **Status:** ✅ Completed - scheduler.go, handlers/regulations.go, routes added to main.go

## Phase 3: Embeddings & Vectorization

- [x] Select embedding provider and implement interface (4h)
  - **Acceptance criteria:** Abstraction interface; provider implementation (e.g., Open‑source sentence transformers via service, or vendor); configurable
  - **Files:** backend/services/embedding/interface.go, backend/services/embedding/provider_xxx.go, backend/config/config.go
  - **Dependencies:** None
  - **Status:** ✅ Completed - interface.go, kolosal.go, factory.go implemented with configurable provider

- [x] Vectorize regulation chunks + upsert to DB (3h)
  - **Acceptance criteria:** Batch job to embed all new chunks; stored in regulation_embeddings with vector index
  - **Files:** backend/services/embedding/indexer.go
  - **Dependencies:** 004/005 migrations, embedding provider
  - **Status:** ✅ Completed - indexer.go implemented with batch processing and pgvector integration

- [x] KNN retrieval API for regulations (2h)
  - **Acceptance criteria:** Service method returning top‑k chunks + metadata; supports filters (year, category)
  - **Files:** backend/services/embedding/retrieval.go
  - **Dependencies:** Indexer
  - **Status:** ✅ Completed - retrieval.go implemented with KNN search, filters, and handlers/embeddings.go with API endpoints

## Phase 4: Chat RAG Integration

- [x] Augment chat SendMessage with retrieval (5h)
  - **Acceptance criteria:** Query embedding → top‑k retrieval from regulations; context window builder; Kolosal.ai completion with citations
  - **Files:** backend/handlers/chat.go (RAG context), backend/handlers/ai.go (shared prompt builder)
  - **Dependencies:** Retrieval service
  - **Status:** ✅ Completed - rag.go service created, SendMessage enhanced with RAG retrieval and context building

- [x] Return citations and source snippets in response (2h)
  - **Acceptance criteria:** API returns sources (title, year, section, URL); UI can render
  - **Files:** backend/handlers/chat.go, frontend/src/components/chat/ChatInterface.tsx
  - **Dependencies:** Chat RAG
  - **Status:** ✅ Completed - Citations added to SendMessageResponse, ExtractCitations function implemented

- [x] Relevance feedback and logging (2h)
  - **Acceptance criteria:** Optional thumbs up/down stored; retrieval diagnostics logged
  - **Files:** backend/handlers/chat.go, database/migrations/009_feedback.sql
  - **Dependencies:** Chat RAG
  - **Status:** ✅ Completed - 009_feedback.sql migration created, handlers/feedback.go implemented, retrieval diagnostics logged

## Phase 5: Google Trends (Scraping, no official API)

- [ ] Trends scraper without official API (6h)
  - **Acceptance criteria:** Headless scraper to fetch interest‑over‑time for keywords/geo; retry/backoff; terms of use compliance
  - **Files:** tools/trends-scraper/ (Node/Playwright or Go/chromedp), README in folder
  - **Dependencies:** 006_trends.sql

- [ ] Backend ingestion + storage (3h)
  - **Acceptance criteria:** Persist time series and related queries; dedup; keyed by keyword and geo
  - **Files:** backend/services/trends/store.go, backend/services/trends/ingest.go
  - **Dependencies:** Scraper artifacts

- [ ] Company‑scoped keyword management endpoints (3h)
  - **Acceptance criteria:** CRUD for tracked keywords per company; list time series and related queries
  - **Files:** backend/handlers/trends.go, backend/main.go (routes)
  - **Dependencies:** Trends storage

## Phase 6: Advanced Forecasting (12 months) + Monthly Strategies

- [ ] Forecasting service (12‑step horizon) (6h)
  - **Acceptance criteria:** Trains on sales_history; includes exogenous signals (trends, seasonality, regulation flags); monthly predictions stored
  - **Files:** services/forecasting/ (Python microservice or Go module), backend/services/forecast/adapter.go
  - **Dependencies:** 007_forecasts_strategies.sql, trends/regulations available

- [ ] Strategy generator per month (4h)
  - **Acceptance criteria:** For each of the 12 months, generate strategy text + structured actions (pricing, inventory, marketing); stored in monthly_strategies
  - **Files:** backend/services/strategy/generator.go
  - **Dependencies:** Forecast outputs

- [ ] API endpoints to retrieve forecasts + strategies (2h)
  - **Acceptance criteria:** `/api/v1/forecasts/monthly` and `/api/v1/strategies/monthly` (company/product scope)
  - **Files:** backend/handlers/forecasts.go (extend), backend/main.go (routes)
  - **Dependencies:** Forecasting service

- [ ] Scheduler for monthly updates + backfill (2h)
  - **Acceptance criteria:** Cron to refresh forecasts monthly; initial backfill for last N months; error alerts
  - **Files:** backend/services/scheduler/scheduler.go
  - **Dependencies:** Forecasting pipeline

## Phase 7: Admin Panel (Separate App) + Billing

- [ ] Create separate admin frontend app (5h)
  - **Acceptance criteria:** New `admin/` app with login, layout, RBAC; build scripts and Dockerfile
  - **Files:** admin/ (new folder), docker-compose.yml (services)
  - **Dependencies:** RBAC schema

- [ ] Backend RBAC and admin endpoints (4h)
  - **Acceptance criteria:** JWT includes role; middleware enforces admin; endpoints: admins CRUD, users CRUD, subscriptions CRUD
  - **Files:** backend/middleware/middleware.go (roles), backend/handlers/admin/*.go, backend/main.go (routes)
  - **Dependencies:** 008_admin_billing_rbac.sql

- [ ] Stripe test mode integration (5h)
  - **Acceptance criteria:** Plans, checkout/test payment, subscription activation/deactivation; webhook handling
  - **Files:** backend/handlers/billing.go, backend/services/billing/stripe.go, backend/config/config.go
  - **Dependencies:** Billing tables

## Phase 8: Security, Observability, Testing

- [ ] Audit logs and rate limiting (3h)
  - **Acceptance criteria:** Sensitive actions logged; rate limits on scraping/trends/admin endpoints
  - **Files:** backend/middleware/*.go, backend/services/audit/logger.go
  - **Dependencies:** Admin endpoints

- [ ] Unit/E2E tests for critical paths (4h)
  - **Acceptance criteria:** Tests for RAG retrieval, trends ingestion, forecasting adapter, admin RBAC; Playwright tests for admin
  - **Files:** backend/*_test.go, admin/tests/e2e/*
  - **Dependencies:** Implementations

## Phase 9: Deployment & DevOps

- [ ] Update Docker Compose and envs (3h)
  - **Acceptance criteria:** New services for scraper, trends, forecasting (if separate); pgvector enabled; env vars documented
  - **Files:** docker-compose.yml, README.md, .env.example
  - **Dependencies:** All services defined

- [ ] Run and verify migrations (1h)
  - **Acceptance criteria:** All new migrations apply cleanly in dev/staging; rollback plan
  - **Files:** database/migrations/*.sql
  - **Dependencies:** Completed migration files

---

## Notes
- Regulations: Do not persist PDFs; only extracted text + metadata + embeddings.
- Google Trends: Use scraping cautiously; cache aggressively; respect site rules.
- Forecasting: Prefer separate service for ML stack; expose adapter from backend.
- Admin: Separate app ensures isolation; enforce strict RBAC; test-mode billing only.


