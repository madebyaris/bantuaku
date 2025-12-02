.PHONY: dev dev-backend dev-frontend build up down logs clean migrate seed test test-backend test-frontend test-e2e

# Use docker compose (v2) or docker-compose (v1)
DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")

# Development
dev: kill-ports
	$(DOCKER_COMPOSE) up --build

# Kill processes using project ports
kill-ports:
	@echo "Checking for processes using ports 3000, 8080, 5432, 6379..."
	@-lsof -ti :3000 | xargs kill -9 2>/dev/null || true
	@-lsof -ti :8080 | xargs kill -9 2>/dev/null || true
	@-lsof -ti :5432 | xargs kill -9 2>/dev/null || true
	@-lsof -ti :6379 | xargs kill -9 2>/dev/null || true
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
	$(DOCKER_COMPOSE) exec db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/002_seed_demo_data.sql

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
	sleep 5
	$(MAKE) migrate
	$(MAKE) seed
