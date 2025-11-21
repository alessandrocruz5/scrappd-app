.PHONY: help dev build test lint migrate-up migrate-down migrate-create seed clean docker-build docker-run

# Variables
BINARY_NAME=scrappd-api
MAIN_PATH=./cmd/server
MIGRATION_PATH=./migrations
DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

# Load environment variables
include .env
export

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# ============================================
# Development
# ============================================

dev: ## Run the application in development mode with hot reload
	@echo "Starting development server with hot reload..."
	air -c .air.toml

run: ## Run the application without hot reload
	@echo "Starting application..."
	go run $(MAIN_PATH)/main.go

build: ## Build the application
	@echo "Building application..."
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go
	@echo "Binary created at bin/$(BINARY_NAME)"

# ============================================
# Testing
# ============================================

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -run Integration ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# ============================================
# Code Quality
# ============================================

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

# ============================================
# Database
# ============================================

migrate-install: ## Install golang-migrate tool
	@echo "Installing golang-migrate..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" -verbose up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" -verbose down

migrate-force: ## Force migration version (use: make migrate-force VERSION=1)
	@echo "Forcing migration version $(VERSION)..."
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(VERSION)

migrate-version: ## Show current migration version
	@echo "Current migration version:"
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" version

migrate-create: ## Create new migration (use: make migrate-create NAME=create_users_table)
	@echo "Creating new migration: $(NAME)..."
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(NAME)

seed: ## Seed the database with sample data
	@echo "Seeding database..."
	go run ./scripts/seed/main.go

# ============================================
# Database Operations
# ============================================

db-connect: ## Connect to database using psql
	@echo "Connecting to database..."
	PGPASSWORD=$(POSTGRES_PASSWORD) psql -h localhost -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-reset: migrate-down migrate-up seed ## Reset database (down, up, seed)
	@echo "Database reset complete!"

db-backup: ## Backup database
	@echo "Backing up database..."
	pg_dump -h localhost -p $(POSTGRES_PORT) -U $(POSTGRES_USER) $(POSTGRES_DB) > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Backup complete!"

# ============================================
# Docker
# ============================================

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t scrappd-api:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env scrappd-api:latest

docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker tag scrappd-api:latest your-registry/scrappd-api:latest
	docker push your-registry/scrappd-api:latest

# ============================================
# Dependencies
# ============================================

deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# ============================================
# Clean
# ============================================

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f *.log
	go clean -cache -testcache -modcache

# ============================================
# Documentation
# ============================================

docs: ## Generate API documentation
	@echo "Generating API documentation..."
	swag init -g $(MAIN_PATH)/main.go -o ./docs

# ============================================
# Production
# ============================================

prod-build: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags='-w -s -extldflags "-static"' \
		-a -installsuffix cgo \
		-o bin/$(BINARY_NAME) \
		$(MAIN_PATH)/main.go
	@echo "Production binary created!"

# ============================================
# CI/CD
# ============================================

ci: lint test ## Run CI pipeline (lint + test)
	@echo "CI pipeline complete!"

# ============================================
# Monitoring
# ============================================

logs: ## Show application logs
	@echo "Showing logs..."
	tail -f logs/app.log

watch: ## Watch for changes and restart (requires entr)
	@echo "Watching for changes..."
	ls **/*.go | entr -r make run