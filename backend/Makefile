
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

# ==============================================
# DOCKER COMPOSE COMMANDS
# ==============================================

db-up: ## Start PostgreSQL and Redis
	@echo "🚀 Starting database services..."
	docker-compose up -d postgres redis
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "✅ Database services started"

db-down: ## Stop database services
	@echo "🛑 Stopping database services..."
	docker-compose down
	@echo "✅ Database services stopped"

db-logs: ## View database logs
	docker-compose logs -f postgres redis

db-ps: ## List running database containers
	docker-compose ps

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U scrappd_app -d scrappd

db-shell-admin: ## Open PostgreSQL shell as admin
	docker-compose exec postgres psql -U postgres -d scrappd

db-health: ## Check if database is ready
	@docker-compose exec postgres pg_isready -U postgres -d scrappd || (echo "❌ Database not ready" && exit 1)
	@echo "✅ Database is ready"

db-reset: db-down ## Reset database (removes all data)
	@echo "⚠️  This will DELETE all data!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || (echo "Aborted" && exit 1)
	docker-compose down -v
	docker-compose up -d postgres redis
	@echo "⏳ Waiting for database..."
	@sleep 5
	@echo "✅ Database reset complete"

db-clean: ## Remove all data from database (keeps structure)
	@echo "⚠️  This will DELETE all data but keep tables!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || (echo "Aborted" && exit 1)
	docker-compose exec postgres psql -U scrappd_admin -d scrappd -c "TRUNCATE auth.users CASCADE;"
	@echo "✅ Database cleaned"

db-backup: ## Backup database to file
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@BACKUP_FILE=backups/scrappd_backup_$$(date +%Y%m%d_%H%M%S).sql; \
	docker-compose exec -T postgres pg_dump -U scrappd_admin scrappd > $$BACKUP_FILE; \
	echo "✅ Backup created: $$BACKUP_FILE"

db-restore: ## Restore database from backup (FILE=path/to/backup.sql)
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Please provide backup file: make db-restore FILE=backups/backup.sql"; \
		exit 1; \
	fi
	@echo "⚠️  This will overwrite current database!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || (echo "Aborted" && exit 1)
	@echo "📦 Restoring database from $(FILE)..."
	@cat $(FILE) | docker-compose exec -T postgres psql -U scrappd_admin scrappd
	@echo "✅ Database restored"

# ML Service
ml-up: ## Start ML service
	docker-compose up -d ml-service

ml-down: ## Stop ML service
	docker-compose stop ml-service

ml-logs: ## View ML service logs
	docker-compose logs -f ml-service

ml-rebuild: ## Rebuild and restart ML service
	docker-compose up -d --build ml-service

# Tools (optional services)
tools-up: ## Start pgAdmin and Redis Commander
	docker-compose --profile tools up -d

tools-down: ## Stop optional tools
	docker-compose --profile tools down

# All services
services-up: ## Start all services (DB + ML)
	docker-compose up -d

services-down: ## Stop all services
	docker-compose down

services-logs: ## View all service logs
	docker-compose logs -f

services-restart: ## Restart all services
	docker-compose restart

# ==============================================
# MIGRATION COMMANDS (updated)
# ==============================================

migrate-up: db-health ## Run database migrations
	@echo "📦 Running migrations..."
	docker-compose exec postgres psql -U scrappd_admin -d scrappd -f /docker-entrypoint-initdb.d/migration.sql
	@echo "✅ Migrations complete"

migrate-create: ## Create new migration file (NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Please provide migration name: make migrate-create NAME=add_users"; \
		exit 1; \
	fi
	@echo "-- Migration: $(NAME)" > ../database/migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).sql
	@echo "✅ Created migration: ../database/migrations/$$(date +%Y%m%d%H%M%S)_$(NAME).sql"

# ==============================================
# SETUP COMMANDS
# ==============================================

setup: install-deps db-up ## Complete first-time setup
	@echo "⏳ Waiting for database to be ready..."
	@sleep 5
	@make db-health
	@echo "✅ Setup complete! Run 'make dev' to start development"

teardown: services-down ## Stop all services and clean up
	docker-compose down -v
	@echo "✅ Teardown complete"

# ==============================================
# JWT SECRET GENERATION
# ==============================================

generate-secrets: ## Generate new JWT secrets
	@echo "🔐 Generating JWT Secrets..."
	@echo ""
	@if command -v openssl > /dev/null; then \
		echo "Access Token Secret:"; \
		openssl rand -base64 32; \
		echo ""; \
		echo "Refresh Token Secret:"; \
		openssl rand -base64 32; \
		echo ""; \
		echo "⚠️  Add these to your .env file:"; \
		echo "JWT_ACCESS_SECRET=<access_token_above>"; \
		echo "JWT_REFRESH_SECRET=<refresh_token_above>"; \
	else \
		echo "⚠️  OpenSSL not found. Using Go script..."; \
		cd ../scripts && $(GO) run generate_secrets.go; \
	fi
	@echo ""

generate-secrets-update: ## Generate and update .env file automatically
	@echo "🔐 Generating and updating JWT secrets..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Create it first from .env.example"; \
		exit 1; \
	fi
	@ACCESS_SECRET=$$(openssl rand -base64 32); \
	REFRESH_SECRET=$$(openssl rand -base64 32); \
	if grep -q "JWT_ACCESS_SECRET" .env; then \
		if [ "$$(uname)" = "Darwin" ]; then \
			sed -i '' "s|JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$$ACCESS_SECRET|" .env; \
			sed -i '' "s|JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$$REFRESH_SECRET|" .env; \
		else \
			sed -i "s|JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$$ACCESS_SECRET|" .env; \
			sed -i "s|JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$$REFRESH_SECRET|" .env; \
		fi; \
		echo "✅ Updated JWT secrets in .env"; \
	else \
		echo "" >> .env; \
		echo "# JWT Secrets" >> .env; \
		echo "JWT_ACCESS_SECRET=$$ACCESS_SECRET" >> .env; \
		echo "JWT_REFRESH_SECRET=$$REFRESH_SECRET" >> .env; \
		echo "JWT_ACCESS_EXPIRY=15m" >> .env; \
		echo "JWT_REFRESH_EXPIRY=168h" >> .env; \
		echo "✅ Added JWT secrets to .env"; \
	fi

# ==============================================
# INTEGRATION TESTS
# ==============================================

test-integration: ## Run integration tests (requires database)
	@echo "🧪 Running integration tests..."
	@if docker-compose ps postgres | grep -q "Up"; then \
		$(GOTEST) -tags=integration ./tests/integration/... -v; \
	else \
		echo "❌ Database not running. Start with: make db-up"; \
		exit 1; \
	fi

test-integration-coverage: ## Run integration tests with coverage
	@echo "🧪 Running integration tests with coverage..."
	@if docker-compose ps postgres | grep -q "Up"; then \
		$(GOTEST) -tags=integration ./tests/integration/... -v -coverprofile=integration_coverage.out; \
		$(GO) tool cover -html=integration_coverage.out -o integration_coverage.html; \
		@echo "Integration coverage report: integration_coverage.html"; \
	else \
		echo "❌ Database not running. Start with: make db-up"; \
		exit 1; \
	fi

test-all: test test-integration ## Run all tests (unit + integration)

# ==============================================
# DEVELOPMENT HELPERS
# ==============================================

fmt: ## Format all Go code
	@echo "🎨 Formatting code..."
	@$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@$(GO) vet ./...
	@echo "✅ Vet complete"

check: fmt vet lint test ## Run all code quality checks

watch-tests: ## Watch and run tests on file changes (requires entr)
	@if command -v entr > /dev/null; then \
		find . -name '*.go' | entr -c make test-unit; \
	else \
		echo "⚠️  entr not found. Install with: brew install entr (macOS) or apt install entr (Linux)"; \
	fi

# ==============================================
# QUICK WORKFLOWS
# ==============================================

fresh-start: teardown setup dev ## Complete teardown and fresh start

dev-full: db-up ml-up dev ## Start all services and run dev

stop-all: services-down ## Stop all services (alias)

# ==============================================
# DOCUMENTATION
# ==============================================

docs: ## Generate API documentation (if using swag)
	@if command -v swag > /dev/null; then \
		swag init -g cmd/api/main.go -o docs; \
		echo "✅ API documentation generated in docs/"; \
	else \
		echo "⚠️  swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# ==============================================
# ENVIRONMENT SETUP
# ==============================================

env-check: ## Check if .env file exists and has required variables
	@echo "🔍 Checking environment configuration..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found"; \
		echo "📝 Copy .env.example to .env and configure it"; \
		exit 1; \
	fi
	@echo "✅ .env file exists"
	@if ! grep -q "JWT_ACCESS_SECRET" .env || grep -q "CHANGE_THIS" .env; then \
		echo "⚠️  JWT secrets not configured. Run: make generate-secrets"; \
	else \
		echo "✅ JWT secrets configured"; \
	fi
	@if ! grep -q "DB_HOST" .env; then \
		echo "⚠️  Database configuration missing"; \
	else \
		echo "✅ Database configuration found"; \
	fi

env-template: ## Create .env from .env.example
	@if [ -f .env ]; then \
		echo "⚠️  .env already exists. Delete it first or use make generate-secrets-update"; \
		exit 1; \
	fi
	@if [ ! -f .env.example ]; then \
		echo "❌ .env.example not found"; \
		exit 1; \
	fi
	@cp .env.example .env
	@echo "✅ Created .env from .env.example"
	@echo "📝 Now run: make generate-secrets-update"

# ==============================================
# PERFORMANCE
# ==============================================

bench: ## Run benchmarks
	@$(GO) test -bench=. -benchmem ./...

profile-cpu: ## Run CPU profiling
	@$(GO) test -cpuprofile=cpu.prof -bench=. ./...
	@echo "View with: go tool pprof cpu.prof"

profile-mem: ## Run memory profiling
	@$(GO) test -memprofile=mem.prof -bench=. ./...
	@echo "View with: go tool pprof mem.prof"
