# Configuration Requirements for RAG + Forecasting + Admin

This document outlines all environment variables, Docker configuration, and setup requirements.

## Environment Variables

### Backend Service

#### Database Configuration
```bash
# PostgreSQL connection (existing)
DATABASE_URL=postgres://bantuaku:bantuaku_secret@db:5432/bantuaku_dev?sslmode=disable

# Note: pgvector extension enabled via migration 004
# No additional env vars needed for pgvector
```

#### Redis Configuration
```bash
# Redis connection (existing)
REDIS_URL=redis://redis:6379
```

#### Authentication & Security
```bash
# JWT secret (existing)
JWT_SECRET=dev-jwt-secret-change-in-production

# CORS origin (existing)
CORS_ORIGIN=http://localhost:3000
```

#### AI/ML Services
```bash
# Kolosal.ai API key (existing, used for chat)
KOLOSAL_API_KEY=your_kolosal_api_key_here

# Embedding provider configuration (NEW)
EMBEDDING_PROVIDER=kolosal  # Options: 'kolosal', 'openai', 'cohere'
EMBEDDING_API_KEY=${KOLOSAL_API_KEY}  # Can use same key or separate

# Forecasting service URL (NEW)
FORECASTING_SERVICE_URL=http://forecasting:8000  # Internal Docker network
# For local dev: http://localhost:8001
```

#### Stripe Billing (NEW)
```bash
# Stripe API keys (test mode initially)
STRIPE_SECRET_KEY=sk_test_...  # Stripe secret key
STRIPE_PUBLISHABLE_KEY=pk_test_...  # Stripe publishable key (frontend)
STRIPE_WEBHOOK_SECRET=whsec_...  # Webhook signing secret

# Stripe configuration
STRIPE_CURRENCY=IDR
STRIPE_MODE=test  # 'test' or 'live'
```

#### Scraper Configuration (NEW)
```bash
# Regulations scraper
REGULATIONS_SCRAPER_ENABLED=true
REGULATIONS_SCRAPER_SCHEDULE=0 2 * * *  # Daily at 2 AM (cron format)
REGULATIONS_BASE_URL=https://peraturan.go.id

# Trends scraper
TRENDS_SCRAPER_ENABLED=true
TRENDS_SCRAPER_SCHEDULE=0 3 * * *  # Daily at 3 AM
TRENDS_GEO_DEFAULT=ID  # Default country code (Indonesia)
TRENDS_RATE_LIMIT_DELAY=2000  # Milliseconds between requests
```

#### Logging & Monitoring
```bash
# Log level (existing)
LOG_LEVEL=debug  # Options: 'debug', 'info', 'warn', 'error'

# Optional: Sentry for error tracking
SENTRY_DSN=your_sentry_dsn_here
```

### Forecasting Service (Python)

```bash
# Database connection (same PostgreSQL)
DATABASE_URL=postgres://bantuaku:bantuaku_secret@db:5432/bantuaku_dev?sslmode=disable

# Service configuration
FORECASTING_PORT=8000
FORECASTING_HOST=0.0.0.0

# Model configuration
FORECASTING_ALGORITHM=prophet  # Options: 'prophet', 'arima', 'lstm'
FORECASTING_MODEL_VERSION=v1.0.0

# Performance
FORECASTING_WORKERS=4  # Number of worker processes
FORECASTING_MAX_CONCURRENT=10  # Max concurrent forecasts
```

### Admin Frontend

```bash
# API endpoint
VITE_API_URL=http://localhost:8080  # Backend API URL

# Admin-specific
VITE_ADMIN_ENABLED=true
VITE_STRIPE_PUBLISHABLE_KEY=${STRIPE_PUBLISHABLE_KEY}
```

## Docker Compose Updates

### PostgreSQL Service (pgvector)

```yaml
services:
  db:
    image: pgvector/pgvector:pg18  # Use pgvector image
    # OR use postgres:18-alpine and install pgvector in migration
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=bantuaku_dev
      - POSTGRES_USER=bantuaku
      - POSTGRES_PASSWORD=bantuaku_secret
    volumes:
      - db_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bantuaku -d bantuaku_dev"]
      interval: 5s
      timeout: 3s
      retries: 5
```

**Note:** Two options for pgvector:
1. Use `pgvector/pgvector:pg18` image (includes pgvector)
2. Use `postgres:18-alpine` and install pgvector via migration (recommended for flexibility)

### Backend Service Updates

```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      # ... existing vars ...
      # NEW variables:
      - EMBEDDING_PROVIDER=${EMBEDDING_PROVIDER:-kolosal}
      - EMBEDDING_API_KEY=${EMBEDDING_API_KEY:-${KOLOSAL_API_KEY}}
      - FORECASTING_SERVICE_URL=http://forecasting:8000
      - STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY:-}
      - STRIPE_PUBLISHABLE_KEY=${STRIPE_PUBLISHABLE_KEY:-}
      - STRIPE_WEBHOOK_SECRET=${STRIPE_WEBHOOK_SECRET:-}
      - STRIPE_CURRENCY=${STRIPE_CURRENCY:-IDR}
      - STRIPE_MODE=${STRIPE_MODE:-test}
      - REGULATIONS_SCRAPER_ENABLED=${REGULATIONS_SCRAPER_ENABLED:-true}
      - REGULATIONS_SCRAPER_SCHEDULE=${REGULATIONS_SCRAPER_SCHEDULE:-0 2 * * *}
      - TRENDS_SCRAPER_ENABLED=${TRENDS_SCRAPER_ENABLED:-true}
      - TRENDS_SCRAPER_SCHEDULE=${TRENDS_SCRAPER_SCHEDULE:-0 3 * * *}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
      forecasting:
        condition: service_healthy  # NEW dependency
```

### Forecasting Service (NEW)

```yaml
services:
  forecasting:
    build:
      context: ./services/forecasting
      dockerfile: Dockerfile
    ports:
      - "8001:8000"  # Expose for local dev
    environment:
      - DATABASE_URL=postgres://bantuaku:bantuaku_secret@db:5432/bantuaku_dev?sslmode=disable
      - FORECASTING_PORT=8000
      - FORECASTING_HOST=0.0.0.0
      - FORECASTING_ALGORITHM=${FORECASTING_ALGORITHM:-prophet}
      - FORECASTING_WORKERS=${FORECASTING_WORKERS:-4}
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### Admin Frontend (NEW)

```yaml
services:
  admin:
    image: node:20-alpine
    working_dir: /app
    ports:
      - "3001:3000"  # Different port from main frontend
    environment:
      - VITE_API_URL=http://backend:8080
      - VITE_ADMIN_ENABLED=true
      - VITE_STRIPE_PUBLISHABLE_KEY=${STRIPE_PUBLISHABLE_KEY:-}
      - CHOKIDAR_USEPOLLING=true
      - DOCKER=true
    depends_on:
      - backend
    volumes:
      - ./admin:/app
      - /app/node_modules
    command: sh -c "npm install && npm run dev -- --host"
    stdin_open: true
    tty: true
```

## Go Configuration Updates

### backend/config/config.go

```go
type Config struct {
    // ... existing fields ...
    
    // Embedding configuration
    EmbeddingProvider string
    EmbeddingAPIKey   string
    
    // Forecasting service
    ForecastingServiceURL string
    
    // Stripe billing
    StripeSecretKey      string
    StripePublishableKey string
    StripeWebhookSecret  string
    StripeCurrency       string
    StripeMode           string
    
    // Scraper configuration
    RegulationsScraperEnabled bool
    RegulationsScraperSchedule string
    RegulationsBaseURL         string
    
    TrendsScraperEnabled bool
    TrendsScraperSchedule string
    TrendsGeoDefault     string
    TrendsRateLimitDelay  int
}

func Load() *Config {
    return &Config{
        // ... existing config ...
        
        EmbeddingProvider: getEnv("EMBEDDING_PROVIDER", "kolosal"),
        EmbeddingAPIKey:   getEnv("EMBEDDING_API_KEY", getEnv("KOLOSAL_API_KEY", "")),
        
        ForecastingServiceURL: getEnv("FORECASTING_SERVICE_URL", "http://localhost:8001"),
        
        StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
        StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
        StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
        StripeCurrency:       getEnv("STRIPE_CURRENCY", "IDR"),
        StripeMode:           getEnv("STRIPE_MODE", "test"),
        
        RegulationsScraperEnabled: getEnvBool("REGULATIONS_SCRAPER_ENABLED", true),
        RegulationsScraperSchedule: getEnv("REGULATIONS_SCRAPER_SCHEDULE", "0 2 * * *"),
        RegulationsBaseURL:         getEnv("REGULATIONS_BASE_URL", "https://peraturan.go.id"),
        
        TrendsScraperEnabled: getEnvBool("TRENDS_SCRAPER_ENABLED", true),
        TrendsScraperSchedule: getEnv("TRENDS_SCRAPER_SCHEDULE", "0 3 * * *"),
        TrendsGeoDefault:     getEnv("TRENDS_GEO_DEFAULT", "ID"),
        TrendsRateLimitDelay:  getEnvInt("TRENDS_RATE_LIMIT_DELAY", 2000),
    }
}

func getEnvBool(key string, defaultValue bool) bool {
    val := os.Getenv(key)
    if val == "" {
        return defaultValue
    }
    return val == "true" || val == "1"
}

func getEnvInt(key string, defaultValue int) int {
    val := os.Getenv(key)
    if val == "" {
        return defaultValue
    }
    i, err := strconv.Atoi(val)
    if err != nil {
        return defaultValue
    }
    return i
}
```

## .env.example Updates

Create/update `.env.example`:

```bash
# Database
DATABASE_URL=postgres://bantuaku:bantuaku_secret@localhost:5432/bantuaku_dev?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=dev-jwt-secret-change-in-production

# CORS
CORS_ORIGIN=http://localhost:3000

# AI Services
KOLOSAL_API_KEY=your_kolosal_api_key_here
EMBEDDING_PROVIDER=kolosal
EMBEDDING_API_KEY=${KOLOSAL_API_KEY}

# Forecasting Service
FORECASTING_SERVICE_URL=http://localhost:8001

# Stripe (Test Mode)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_CURRENCY=IDR
STRIPE_MODE=test

# Scrapers
REGULATIONS_SCRAPER_ENABLED=true
REGULATIONS_SCRAPER_SCHEDULE=0 2 * * *
REGULATIONS_BASE_URL=https://peraturan.go.id

TRENDS_SCRAPER_ENABLED=true
TRENDS_SCRAPER_SCHEDULE=0 3 * * *
TRENDS_GEO_DEFAULT=ID
TRENDS_RATE_LIMIT_DELAY=2000

# Logging
LOG_LEVEL=debug
```

## pgvector Setup Notes

### Option 1: Use pgvector Docker Image (Recommended)

```yaml
db:
  image: pgvector/pgvector:pg18
```

**Pros:**
- pgvector pre-installed
- No migration needed for extension
- Easier setup

**Cons:**
- Less control over PostgreSQL version
- Larger image size

### Option 2: Install pgvector in Migration

```yaml
db:
  image: postgres:18-alpine
```

Migration 004 includes:
```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

**Pros:**
- More control over PostgreSQL version
- Smaller base image
- Explicit extension management

**Cons:**
- Requires pgvector available in container (may need custom Dockerfile)

### Recommended Approach

Use `postgres:18-alpine` with custom Dockerfile that installs pgvector:

```dockerfile
FROM postgres:18-alpine

RUN apk add --no-cache \
    git \
    build-base \
    postgresql-dev

# Install pgvector
RUN git clone --branch v0.5.1 https://github.com/pgvector/pgvector.git /tmp/pgvector && \
    cd /tmp/pgvector && \
    make && \
    make install && \
    rm -rf /tmp/pgvector

# Extension will be enabled via migration
```

Or use the official pgvector image for simplicity.

## Python Forecasting Service Dependencies

### services/forecasting/requirements.txt

```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
psycopg2-binary==2.9.9
pandas==2.1.3
numpy==1.26.2
prophet==1.1.5
scikit-learn==1.3.2
statsmodels==0.14.0
httpx==0.25.1
```

### services/forecasting/Dockerfile

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Expose port
EXPOSE 8000

# Run service
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Verification Checklist

After configuration:

- [ ] PostgreSQL has pgvector extension enabled (`SELECT * FROM pg_extension WHERE extname = 'vector';`)
- [ ] Backend can connect to database
- [ ] Backend can connect to Redis
- [ ] Backend can connect to forecasting service
- [ ] Embedding provider API key works
- [ ] Stripe test mode keys configured
- [ ] Scraper schedules configured correctly
- [ ] All environment variables loaded in Go config
- [ ] Docker Compose services start successfully
- [ ] Health checks pass for all services

## Production Considerations

1. **Secrets Management**: Use secret management (AWS Secrets Manager, HashiCorp Vault)
2. **Environment Separation**: Different `.env` files for dev/staging/prod
3. **pgvector Tuning**: Monitor and tune ivfflat index parameters
4. **Rate Limiting**: Implement rate limits for scraping endpoints
5. **Monitoring**: Set up alerts for scraper failures, API errors
6. **Backup**: Regular backups of PostgreSQL (including vector data)

