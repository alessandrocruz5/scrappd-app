
.PHONY: help run test test-coverage build clean dev lint install-tools install-air

# Load environment variables from .env file
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Variables
BINARY_NAME=scrappd-api
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
PROJECT_ROOT := $(shell pwd)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

install-air: ## Install air for hot reload
	@echo "Installing air..."
	@go install github.com/air-verse/air@latest
	@echo "✓ Air installed successfully"
	@echo "Make sure $(shell go env GOPATH)/bin is in your PATH"

install-tools: install-air ## Install all development tools
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "✓ All tools installed"

run: ## Run the application
	$(GO) run cmd/api/main.go

dev: ## Run with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "⚠️  Air not found. Installing..."; \
		make install-air; \
		echo ""; \
		echo "✓ Air installed. Please run 'make dev' again"; \
		echo "  Or add $(shell go env GOPATH)/bin to your PATH and retry"; \
	fi

dev-manual: ## Run with manual reload (no air needed)
	$(GO) run cmd/api/main.go

build: ## Build the application
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/api/main.go

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	$(GOTEST) -v -short ./...

test-handlers: ## Run handler tests
	$(GOTEST) -v ./internal/api/handlers/...

test-services: ## Run service tests
	$(GOTEST) -v ./internal/services/...

test-utils: ## Run utils tests
	$(GOTEST) -v ./pkg/utils/...

lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install with: make install-tools"; \
	fi

clean: ## Clean build files
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f integration_coverage.out integration_coverage.html
	rm -f test_result.png test_result_base64.png
	rm -f cpu.prof mem.prof
	rm -rf tmp/

docker-build: ## Build Docker image
	docker build -t scrappd-api:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env scrappd-api:latest

# Quick development commands
quick-start: install-deps dev-manual ## Quick start without air
dev-with-air: install-air dev ## Install air and run with hot reload

# Database Migration Commands

.PHONY: migrate-up migrate-down migrate-force migrate-create

# Simple migration commands with hardcoded values for local dev
migrate-up-local:
	@echo "Running migrations..."
	migrate -path migrations -database "postgresql://apa@localhost:5432/scrappd?sslmode=disable" up

migrate-status-local:
	@echo "Checking migration status..."
	migrate -path migrations -database "postgresql://apa@localhost:5432/scrappd?sslmode=disable" version

migrate-down-local:
	@echo "Rolling back last migration..."
	migrate -path migrations -database "postgresql://apa@localhost:5432/scrappd?sslmode=disable" down 1

# Run all pending migrations
migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

# Rollback the last migration
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

# Force migration to specific version (use with caution)
migrate-force:
	@echo "Forcing migration to version $(VERSION)..."
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" force $(VERSION)

# Create a new migration file
migrate-create:
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

# Check migration status
migrate-status:
	@echo "Checking migration status..."
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" version

# Repository Tests

.PHONY: test-repo test-repo-verbose test-repo-coverage

# Run repository tests
test-repo:
	@echo "Running repository tests..."
	go test -v ./internal/repository/postgres/... -count=1

# Run repository tests with verbose output
test-repo-verbose:
	@echo "Running repository tests (verbose)..."
	go test -v -race ./internal/repository/postgres/... -count=1

# Run repository tests with coverage
test-repo-coverage:
	@echo "Running repository tests with coverage..."
	go test -v -race -coverprofile=coverage_repo.out -covermode=atomic ./internal/repository/postgres/...
	@echo "Coverage report saved to coverage_repo.out"
	@echo "View coverage report: go tool cover -html=coverage_repo.out"

# Complete Test Suite

.PHONY: test-all test-integration

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests..."
	go test -v -race ./... -count=1

# Run integration tests only (requires services to be running)
test-integration:
	@echo "Running integration tests..."
	go test -v -race -tags=integration ./test/integration/... -count=1
