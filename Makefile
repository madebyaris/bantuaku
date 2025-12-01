.PHONY: dev dev-backend dev-frontend build up down logs clean migrate seed

# Use docker compose (v2) or docker-compose (v1)
DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")

# Development
dev:
	$(DOCKER_COMPOSE) up --build

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

seed:
	$(DOCKER_COMPOSE) exec db psql -U bantuaku -d bantuaku_dev -f /docker-entrypoint-initdb.d/002_seed_demo_data.sql

# Cleanup
clean:
	$(DOCKER_COMPOSE) down -v
	rm -rf frontend/node_modules
	rm -rf frontend/dist

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

# Full reset
reset: clean
	$(DOCKER_COMPOSE) up --build -d
	sleep 5
	$(MAKE) migrate
	$(MAKE) seed
