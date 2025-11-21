.PHONY: help setup start stop restart clean logs test lint format

# Default target
help: ## Show this help message
	@echo 'Scrapp'"'"'d - Digital Scrapbooking Platform'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# ============================================
# Setup
# ============================================

setup: ## Run initial setup script
	@echo "Setting up Scrapp'd development environment..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# ============================================
# Docker Compose
# ============================================

start: ## Start all services
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Services started!"
	@echo "API:        http://localhost:8080"
	@echo "ML Service: http://localhost:8000"
	@echo "ML Docs:    http://localhost:8000/docs"

stop: ## Stop all services
	@echo "Stopping all services..."
	docker-compose down

restart: ## Restart all services
	@echo "Restarting all services..."
	docker-compose restart

start-infra: ## Start only infrastructure services (postgres, redis)
	@echo "Starting infrastructure services..."
	docker-compose up -d postgres redis

stop-infra: ## Stop infrastructure services
	@echo "Stopping infrastructure services..."
	docker-compose stop postgres redis

start-tools: ## Start admin tools (pgadmin, redis-commander)
	@echo "Starting admin tools..."
	docker-compose --profile tools up -d
	@echo "Tools started!"
	@echo "pgAdmin:         http://localhost:5050"
	@echo "Redis Commander: http://localhost:8081"

# ============================================
# Development
# ============================================

dev-api: ## Start API service in development mode
	@echo "Starting API service..."
	cd services/api && make dev

dev-ml: ## Start ML service in development mode
	@echo "Starting ML service..."
	cd services/ml-service && make dev

dev-mobile: ## Start mobile app
	@echo "Starting mobile app..."
	cd mobile && flutter run

# ============================================
# Database
# ============================================

db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	cd services/api && make migrate-up

db-reset: ## Reset database (WARNING: destructive)
	@echo "Resetting database..."
	cd services/api && make db-reset

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	cd services/api && make seed

db-backup: ## Backup database
	@echo "Backing up database..."
	cd services/api && make db-backup

# ============================================
# Testing
# ============================================

test: ## Run all tests
	@echo "Running all tests..."
	@$(MAKE) test-api
	@$(MAKE) test-ml
	@$(MAKE) test-mobile

test-api: ## Run API tests
	@echo "Running API tests..."
	cd services/api && make test

test-ml: ## Run ML service tests
	@echo "Running ML service tests..."
	cd services/ml-service && make test

test-mobile: ## Run mobile tests
	@echo "Running mobile tests..."
	cd mobile && flutter test

# ============================================
# Code Quality
# ============================================

lint: ## Run linters on all services
	@echo "Running linters..."
	@$(MAKE) lint-api
	@$(MAKE) lint-ml
	@$(MAKE) lint-mobile

lint-api: ## Lint API code
	@echo "Linting API code..."
	cd services/api && make lint

lint-ml: ## Lint ML service code
	@echo "Linting ML service code..."
	cd services/ml-service && make lint

lint-mobile: ## Lint mobile code
	@echo "Linting mobile code..."
	cd mobile && flutter analyze

format: ## Format all code
	@echo "Formatting code..."
	@$(MAKE) format-api
	@$(MAKE) format-ml

format-api: ## Format API code
	@echo "Formatting API code..."
	cd services/api && make fmt

format-ml: ## Format ML service code
	@echo "Formatting ML service code..."
	cd services/ml-service && make format

# ============================================
# Build
# ============================================

build: ## Build all services
	@echo "Building all services..."
	docker-compose build

build-api: ## Build API service
	@echo "Building API service..."
	cd services/api && make build

build-ml: ## Build ML service
	@echo "Building ML service..."
	docker-compose build ml-service

build-mobile: ## Build mobile app
	@echo "Building mobile app..."
	cd mobile && flutter build apk

# ============================================
# Logs
# ============================================

logs: ## Show logs from all services
	docker-compose logs -f

logs-api: ## Show API service logs
	docker-compose logs -f api

logs-ml: ## Show ML service logs
	docker-compose logs -f ml-service

logs-postgres: ## Show PostgreSQL logs
	docker-compose logs -f postgres

logs-redis: ## Show Redis logs
	docker-compose logs -f redis

# ============================================
# Clean
# ============================================

clean: ## Clean all build artifacts and cache
	@echo "Cleaning..."
	@$(MAKE) clean-api
	@$(MAKE) clean-ml
	@$(MAKE) clean-mobile
	@$(MAKE) clean-docker

clean-api: ## Clean API artifacts
	cd services/api && make clean

clean-ml: ## Clean ML service artifacts
	cd services/ml-service && make clean

clean-mobile: ## Clean mobile artifacts
	cd mobile && flutter clean

clean-docker: ## Clean Docker containers and volumes
	@echo "Cleaning Docker containers and volumes..."
	docker-compose down -v

# ============================================
# Documentation
# ============================================

docs: ## Generate documentation
	@echo "Generating documentation..."
	cd services/api && make docs

# ============================================
# CI/CD
# ============================================

ci: lint test ## Run CI pipeline (lint + test)
	@echo "CI pipeline complete!"

# ============================================
# Utilities
# ============================================

ps: ## Show running containers
	docker-compose ps

shell-api: ## Open shell in API container
	docker-compose exec api sh

shell-ml: ## Open shell in ML service container
	docker-compose exec ml-service bash

shell-postgres: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U scrappd -d scrappd

shell-redis: ## Open Redis CLI
	docker-compose exec redis redis-cli -a scrappd_redis_password

download-models: ## Download ML models
	cd services/ml-service && make download-models

# ============================================
# Status
# ============================================

status: ## Show status of all services
	@echo "=== Scrapp'd Services Status ==="
	@echo ""
	@docker-compose ps
	@echo ""
	@echo "=== URLs ==="
	@echo "API:              http://localhost:8080"
	@echo "ML Service:       http://localhost:8000"
	@echo "ML Docs:          http://localhost:8000/docs"
	@echo "pgAdmin:          http://localhost:5050"
	@echo "Redis Commander:  http://localhost:8081"