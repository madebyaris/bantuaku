# All-in-One Dockerfile for Railway
# Runs: Backend (Go) + Frontend (React/Nginx) + Forecasting (Python)

# ============================================
# Stage 1: Build Go Backend
# ============================================
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app
RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bantuaku .

# ============================================
# Stage 2: Build React Frontend
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --legacy-peer-deps

COPY frontend/ .
RUN npm run build

# ============================================
# Stage 3: Build Python Forecasting Dependencies
# ============================================
FROM python:3.11-slim AS forecasting-builder

WORKDIR /app
COPY services/forecasting/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt --target=/app/deps

# ============================================
# Stage 4: Final Runtime Image
# ============================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    poppler-utils \
    nginx \
    python3 \
    py3-pip \
    supervisor

WORKDIR /app

# Copy Go backend binary
COPY --from=backend-builder /app/bantuaku /app/bantuaku

# Copy Frontend build to nginx
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/http.d/default.conf

# Copy Python forecasting service
COPY --from=forecasting-builder /app/deps /app/forecasting/deps
COPY services/forecasting/app /app/forecasting/app

# Create supervisord config to run all services
RUN mkdir -p /var/log/supervisor
COPY <<EOF /etc/supervisord.conf
[supervisord]
nodaemon=true
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid
user=root

[program:backend]
command=/app/bantuaku
directory=/app
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
environment=PORT="8080"

[program:nginx]
command=nginx -g "daemon off;"
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:forecasting]
command=python3 -m uvicorn app.main:app --host 0.0.0.0 --port 8000
directory=/app/forecasting
environment=PYTHONPATH="/app/forecasting/deps"
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

# Expose ports
# Railway will use PORT env var, but we expose these for documentation
EXPOSE 8080 3000 8000

# Health check on backend
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/api/v1/health || exit 1

# Run supervisord to manage all processes
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
