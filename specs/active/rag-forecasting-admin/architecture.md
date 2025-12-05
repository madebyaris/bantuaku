# RAG + Forecasting + Admin Panel - System Architecture

## Overview

This document describes the architecture for three major features:
1. **RAG (Retrieval-Augmented Generation)** - Regulations knowledge base with vector search
2. **12-Month Forecasting** - Advanced forecasting with monthly strategies
3. **Admin Panel** - Separate admin application with RBAC and billing

## System Context

```
┌─────────────────────────────────────────────────────────────┐
│                    External Systems                         │
├─────────────────────────────────────────────────────────────┤
│  peraturan.go.id  │  Google Trends  │  Stripe (Billing)    │
└─────────────────────────────────────────────────────────────┘
                        │         │            │
                        ▼         ▼            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Bantuaku Backend (Go)                   │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ Regulations  │  │   Trends     │  │ Forecasting  │    │
│  │   Scraper    │  │   Scraper    │  │   Service    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ Embeddings  │  │   RAG Chat    │  │   Admin     │    │
│  │   Service   │  │   Handler     │  │   RBAC      │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
                        │         │            │
                        ▼         ▼            ▼
┌─────────────────────────────────────────────────────────────┐
│              PostgreSQL (pgvector) + Redis                  │
└─────────────────────────────────────────────────────────────┘
                        │         │
                        ▼         ▼
┌─────────────────────────────────────────────────────────────┐
│         Frontend Apps (React + Vite + Tailwind)            │
├─────────────────────────────────────────────────────────────┤
│  Main App (User)  │  Admin App (Separate)                  │
└─────────────────────────────────────────────────────────────┘
```

## Container Architecture

### Backend Services

1. **Main API Server** (`backend/`)
   - HTTP handlers for all endpoints
   - Middleware (auth, logging, CORS)
   - Service orchestration

2. **Regulations Scraper** (`backend/services/scraper/regulations/`)
   - Crawler: Enumerate PDFs from peraturan.go.id
   - Extractor: PDF text extraction (temp download, discard)
   - Chunker: Semantic chunking with overlap
   - Store: Dedup and versioning

3. **Embeddings Service** (`backend/services/embedding/`)
   - Provider abstraction (vendor API)
   - Vectorization pipeline
   - KNN retrieval

4. **Trends Scraper** (`tools/trends-scraper/`)
   - Headless browser scraping (chromedp/Playwright)
   - Time series extraction
   - Related queries capture

5. **Forecasting Service** (`services/forecasting/` - Python microservice)
   - 12-month horizon predictions
   - Exogenous signals integration
   - Monthly strategy generation

6. **Admin Backend** (`backend/handlers/admin/`)
   - RBAC enforcement
   - User/Admin CRUD
   - Subscription management
   - Stripe webhook handling

### Frontend Applications

1. **Main App** (`frontend/`)
   - User-facing dashboard
   - Chat interface with RAG
   - Forecast visualization
   - Market trends display

2. **Admin App** (`admin/` - new)
   - Separate React app
   - Admin dashboard
   - User management
   - Subscription management
   - Audit logs viewer

## Data Flow Diagrams

### Regulations RAG Flow

```
1. Crawl peraturan.go.id
   └─> Discover PDF links + metadata
       └─> Store in regulations, regulation_sources

2. Extract PDF Text
   └─> Download to temp
       └─> Extract text (OCR fallback if needed)
           └─> Store raw text in regulation_sections
               └─> Discard PDF

3. Chunk Text
   └─> Split into semantic chunks (overlap)
       └─> Store in regulation_chunks

4. Generate Embeddings
   └─> Batch embed chunks
       └─> Store vectors in regulation_embeddings (pgvector)

5. Chat Query Flow
   └─> User question → Embed query
       └─> KNN search (cosine similarity)
           └─> Top-k chunks retrieved
               └─> Build context window
                   └─> Kolosal.ai completion with citations
                       └─> Return response + sources
```

### Google Trends Flow

```
1. Scrape Trends
   └─> Headless browser → Google Trends
       └─> Extract interest-over-time
           └─> Extract related queries
               └─> Store in trends_keywords, trends_series, trends_related_queries

2. Company Keyword Management
   └─> Admin/User adds keywords
       └─> Scheduler runs daily
           └─> Fetch latest trends
               └─> Upsert time series

3. API Exposure
   └─> GET /api/v1/trends/keywords
       └─> GET /api/v1/trends/series?keyword_id=X
           └─> Return time series + related queries
```

### Forecasting Flow

```
1. Data Collection
   └─> Sales history (sales_history table)
       └─> Google Trends signals (trends_series)
           └─> Regulation flags (regulation_chunks relevance)
               └─> Aggregate into forecast_inputs

2. Forecasting Service (Python)
   └─> Train time-series model
       └─> Generate 12-month predictions
           └─> Store in forecasts_monthly

3. Strategy Generation
   └─> For each month (1-12)
       └─> Generate strategy text (reasoning)
           └─> Extract structured actions (pricing, inventory, marketing)
               └─> Store in monthly_strategies

4. API Exposure
   └─> GET /api/v1/forecasts/monthly?product_id=X
       └─> GET /api/v1/strategies/monthly?product_id=X
           └─> Return forecasts + strategies
```

### Admin Panel Flow

```
1. Authentication
   └─> Admin login → JWT with role claim
       └─> RBAC middleware validates admin role

2. User Management
   └─> CRUD operations on users table
       └─> Audit logs for sensitive actions

3. Subscription Management
   └─> Stripe checkout → Webhook
       └─> Update subscriptions table
           └─> Activate/deactivate features

4. Admin Dashboard
   └─> Aggregate metrics
       └─> System health monitoring
           └─> Recent audit logs
```

## Component Details

### Database Schema (PostgreSQL + pgvector)

**Core Tables:**
- `regulations` - Regulation metadata
- `regulation_sections` - Raw text sections
- `regulation_chunks` - Chunked text
- `regulation_embeddings` - Vector embeddings (pgvector)
- `trends_keywords` - Tracked keywords
- `trends_series` - Time series data
- `forecasts_monthly` - 12-month predictions
- `monthly_strategies` - Strategy per month
- `subscription_plans` - Billing plans
- `subscriptions` - User subscriptions
- `audit_logs` - Security audit trail

**Indexes:**
- Vector index: `regulation_embeddings.embedding` (ivfflat, cosine)
- Time series: `trends_series(keyword_id, time)`
- Forecasts: `forecasts_monthly(product_id, month)`

### Service Interfaces

**Embedding Service:**
```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
```

**Retrieval Service:**
```go
type Retriever interface {
    Search(ctx context.Context, queryEmbedding []float32, k int, filters Filters) ([]Chunk, error)
}
```

**Forecasting Adapter:**
```go
type ForecastAdapter interface {
    GenerateForecast(ctx context.Context, inputs ForecastInputs) (*Forecast, error)
    GenerateStrategies(ctx context.Context, forecast *Forecast) ([]MonthlyStrategy, error)
}
```

## Technology Stack Decisions

See ADRs for detailed rationale:
- **ADR-001**: pgvector (PostgreSQL extension) for vector storage
- **ADR-002**: Go-based scraper with chromedp
- **ADR-003**: Vendor API for embeddings (configurable provider)
- **ADR-004**: Python microservice for forecasting (FastAPI)

## Deployment Architecture

### Docker Compose Services

```
services:
  backend:          # Main Go API
  frontend:         # Main React app
  admin:            # Admin React app (new)
  db:               # PostgreSQL 18 + pgvector
  redis:            # Caching layer
  forecasting:      # Python FastAPI service (new)
```

### Environment Variables

See `config-requirements.md` for complete list.

Key variables:
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `EMBEDDING_PROVIDER` - Embedding service (e.g., "kolosal", "openai")
- `EMBEDDING_API_KEY` - Provider API key
- `FORECASTING_SERVICE_URL` - Python service endpoint
- `STRIPE_SECRET_KEY` - Billing integration

## Security Considerations

1. **RBAC**: JWT claims include role; middleware enforces admin endpoints
2. **Rate Limiting**: Applied to scraping endpoints and admin operations
3. **Audit Logging**: All sensitive actions logged with user_id, timestamp, action
4. **Data Privacy**: PDFs not persisted; only extracted text stored
5. **Scraping Ethics**: Respect robots.txt, rate limits, ToS compliance

## Scalability Considerations

1. **Vector Search**: pgvector ivfflat index for efficient KNN
2. **Caching**: Redis for forecast results (1h TTL), trends data (6h TTL)
3. **Batch Processing**: Embeddings generated in batches, not per-request
4. **Async Jobs**: Scraping and forecasting run as background jobs
5. **Horizontal Scaling**: Stateless backend services can scale horizontally

## Monitoring & Observability

1. **Structured Logging**: JSON logs with request IDs
2. **Health Checks**: `/healthz` endpoint for all services
3. **Metrics**: Forecast accuracy, retrieval latency, scraping success rate
4. **Error Tracking**: Centralized error handling with context

## Migration Path

See `migrations-outline.md` for detailed migration plan.

Phases:
1. **Phase 1**: Database foundation (pgvector, core tables)
2. **Phase 2**: Regulations scraper + embeddings
3. **Phase 3**: RAG integration in chat
4. **Phase 4**: Trends scraper
5. **Phase 5**: Forecasting service
6. **Phase 6**: Admin panel + billing

## Next Steps

1. Review ADRs for technology decisions
2. Review migration outlines for database schema
3. Review config requirements for environment setup
4. Proceed to Phase 1 implementation

