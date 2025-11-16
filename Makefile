.PHONY: help build up down logs clean test lint fmt db-logs db-shell setup-env

help:
	@echo "=== PR Reviewer Service - Makefile ==="
	@echo ""
	@echo "Available commands:"
	@echo "  make setup-env    - Create .env from .env.example"
	@echo "  make build        - Build Docker images"
	@echo "  make up           - Start all containers (docker-compose up -d)"
	@echo "  make down         - Stop all containers (docker-compose down)"
	@echo "  make logs         - Show logs from all containers"
	@echo "  make logs-app     - Show logs from app container"
	@echo "  make logs-db      - Show logs from db container"
	@echo "  make clean        - Remove containers and volumes"
	@echo "  make db-shell     - Connect to PostgreSQL shell"
	@echo "  make db-logs      - Show database logs"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make dev          - Build and run in development mode"
	@echo "  make rebuild      - Rebuild and restart app container (after code changes)"
	@echo "  make restart      - Quick restart app container"
	@echo "  make fresh        - Complete rebuild from scratch"

# Setup environment
setup-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "✓ .env created. Please update it with your actual values."; \
	else \
		echo ".env already exists. Skipping..."; \
	fi

# Docker commands
build: setup-env
	@echo "Building Docker images..."
	docker-compose build

up: setup-env
	@echo "Starting containers..."
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	docker-compose ps

down:
	@echo "Stopping containers..."
	docker-compose down

restart: down up

logs:
	docker-compose logs -f

logs-app:
	docker-compose logs -f backend-app

logs-db:
	docker-compose logs -f db

clean:
	@echo "Cleaning up..."
	docker-compose down -v
	@echo "Note: .env file preserved. Use 'rm .env' to remove it manually."

# Database commands
db-shell:
	@docker exec -it postgres_db psql -U postgres -d pr_reviewer_db

db-logs:
	docker-compose logs db

# Testing and code quality
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Development
dev: setup-env build up
	@echo "✓ Development environment ready"
	@echo "API available at http://localhost:8080"
	@echo "Database: postgres://localhost:5432/pr_reviewer_db"

# Health check
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "Service not available"

# Database status
db-status:
	@docker exec postgres_db pg_isready -U postgres || echo "Database not ready"

# Show environment
env:
	@echo "Current environment:"
	@cat .env 2>/dev/null || cat .env.example

# Rebuild and restart only app container
rebuild:
	docker-compose build backend-app && docker-compose up -d backend-app

# Restart app container (quick)
restart:
	docker-compose restart backend-app

# Complete rebuild: stop, build, start
fresh:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Development mode with rebuild
dev:
	docker-compose up -d --build backend-app

# Generate OpenAPI code
generate:
	oapi-codegen -generate types      -o api/types.gen.go   -package api api/openapi.yml
	oapi-codegen -generate chi-server -o api/server.gen.go  -package api api/openapi.yml
	oapi-codegen -generate client     -o api/client.gen.go  -package api api/openapi.yml