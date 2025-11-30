.PHONY: help run test test-coverage build clean dev lint install-tools

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

run: ## Run the application
	$(GO) run cmd/api/main.go

dev: ## Run with hot reload (requires air)
	air

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

lint: ## Run linter
	golangci-lint run

clean: ## Clean build files
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t scrappd-api:latest .

docker-run: ## Run Docker container
	docker run -p 8080: