.PHONY: help run test test-coverage build clean dev lint install-tools install-air

# Variables
BINARY_NAME=scrappd-api
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean

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
	rm -f test_result.png test_result_base64.png

docker-build: ## Build Docker image
	docker build -t scrappd-api:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env scrappd-api:latest

# Quick development commands
quick-start: install-deps dev-manual ## Quick start without air
dev-with-air: install-air dev ## Install air and run with hot reload