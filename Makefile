.PHONY: dev dev-backend dev-frontend build up down logs clean migrate seed test test-backend test-frontend test-e2e kill-ports

# Detect OS and set Docker Compose command
ifeq ($(OS),Windows_NT)
    # Windows - use docker compose (v2) directly
    DOCKER_COMPOSE := docker compose
    KILL_PORTS_CMD := @echo "Port checking skipped on Windows. Please manually stop processes if needed."
else
    # Unix/Linux/macOS - try docker-compose first, fallback to docker compose
    DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")
    KILL_PORTS_CMD := @-lsof -ti :3000 | xargs kill -9 2>/dev/null || true; \
	                  -lsof -ti :8080 | xargs kill -9 2>/dev/null || true; \
	                  -lsof -ti :5432 | xargs kill -9 2>/dev/null || true; \
	                  -lsof -ti :6379 | xargs kill -9 2>/dev/null || true
endif

# Development
dev: kill-ports
	$(DOCKER_COMPOSE) up --build

# Kill processes using project ports (optional, skips on Windows)
kill-ports:
	@echo "Checking for processes using ports 3000, 8080, 5432, 6379..."
	$(KILL_PORTS_CMD)
	@echo "Ports cleared"

dev-backend:
	cd backend && go run main.go

dev-frontend:
	cd frontend && npm run dev

# Docker
build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up -d

up-build:
	$(DOCKER_COMPOSE) up --build -d

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f

logs-backend:
	$(DOCKER_COMPOSE) logs -f backend

logs-frontend:
	$(DOCKER_COMPOSE) logs -f frontend

# Database
migrate:
	$(DOCKER_COMPOSE) exec db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/001_init_schema.sql

migrate-remove-stock:
	$(DOCKER_COMPOSE) exec db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/003_remove_stock.sql

seed:
	$(DOCKER_COMPOSE) exec -T db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/010_seed_demo_data.sql

seed-admin:
	@echo "Creating super admin user..."
	$(DOCKER_COMPOSE) exec -T db psql -U bantuaku -d bantuaku_dev -c "INSERT INTO users (id, email, password_hash, role, created_at) VALUES ('super-admin-001', 'admin@bantuaku.id', '\$$2a\$$10\$$E/KmS9sT76xcwUeji.gEDeikxK99miVSTZ9XCLrzcLYayVzvMT1JK', 'super_admin', NOW()) ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = 'super_admin';"
	@echo "Super admin created: admin@bantuaku.id / demo123"

# Cleanup
clean:
	$(DOCKER_COMPOSE) down -v
	rm -rf frontend/node_modules
	rm -rf frontend/dist

# Remove database volume (fixes PostgreSQL upgrade issues)
clean-db:
	$(DOCKER_COMPOSE) down -v
	docker volume rm bantuaku_db_data 2>/dev/null || true

# Install dependencies
install:
	cd frontend && npm install
	cd backend && go mod download

# Go commands
go-tidy:
	cd backend && go mod tidy

go-test:
	cd backend && go test ./...

# Frontend commands
npm-install:
	cd frontend && npm install

npm-build:
	cd frontend && npm run build

# Testing
test:
	$(MAKE) test-backend && $(MAKE) test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

test-e2e:
	cd frontend && npm run test:e2e

test-coverage:
	cd backend && go test -coverprofile=coverage.out ./...
	cd frontend && npm run test:coverage

# Full reset
reset: clean
	$(DOCKER_COMPOSE) up --build -d
ifeq ($(OS),Windows_NT)
	@timeout /t 5 /nobreak >nul 2>&1 || ping 127.0.0.1 -n 6 >nul
else
	@sleep 5
endif
	$(MAKE) migrate
	$(MAKE) seed
