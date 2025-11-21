.PHONY: help dev install test lint format clean docker-build docker-run download-models

# Variables
PYTHON=python3
PIP=pip3
PYTEST=pytest
BLACK=black
FLAKE8=flake8
ISORT=isort
MYPY=mypy

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
	uvicorn src.api.main:app --host 0.0.0.0 --port 8000 --reload

run: ## Run the application in production mode
	@echo "Starting application..."
	gunicorn src.api.main:app \
		--workers 4 \
		--worker-class uvicorn.workers.UvicornWorker \
		--bind 0.0.0.0:8000 \
		--timeout 120

# ============================================
# Installation
# ============================================

install: ## Install dependencies
	@echo "Installing dependencies..."
	$(PIP) install -r requirements.txt

install-dev: ## Install development dependencies
	@echo "Installing development dependencies..."
	$(PIP) install -r requirements.txt
	$(PIP) install -r requirements-dev.txt

venv: ## Create virtual environment
	@echo "Creating virtual environment..."
	$(PYTHON) -m venv venv
	@echo "Activate with: source venv/bin/activate (Linux/Mac) or venv\Scripts\activate (Windows)"

# ============================================
# Testing
# ============================================

test: ## Run all tests
	@echo "Running tests..."
	$(PYTEST) tests/ -v --cov=src --cov-report=html --cov-report=term

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	$(PYTEST) tests/unit -v

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	$(PYTEST) tests/integration -v

test-watch: ## Run tests in watch mode
	@echo "Running tests in watch mode..."
	$(PYTEST) tests/ -v --watch

coverage: ## Generate coverage report
	@echo "Generating coverage report..."
	$(PYTEST) tests/ --cov=src --cov-report=html
	@echo "Coverage report generated at htmlcov/index.html"

# ============================================
# Code Quality
# ============================================

lint: ## Run linter
	@echo "Running flake8..."
	$(FLAKE8) src tests
	@echo "Running mypy..."
	$(MYPY) src

format: ## Format code with black and isort
	@echo "Formatting code with black..."
	$(BLACK) src tests
	@echo "Sorting imports with isort..."
	$(ISORT) src tests

format-check: ## Check code formatting
	@echo "Checking code formatting..."
	$(BLACK) --check src tests
	$(ISORT) --check-only src tests

# ============================================
# Model Management
# ============================================

download-models: ## Download pre-trained models
	@echo "Downloading pre-trained models..."
	@mkdir -p models
	@echo "Downloading RMBG-2.0..."
	@# Add actual download commands here
	@echo "Models downloaded to ./models/"

download-rmbg: ## Download RMBG-2.0 model
	@echo "Downloading RMBG-2.0 model..."
	@mkdir -p models/rmbg
	wget -O models/rmbg/model.pth https://huggingface.co/briaai/RMBG-2.0/resolve/main/model.pth

download-u2net: ## Download U2-Net model
	@echo "Downloading U2-Net model..."
	@mkdir -p models/u2net
	wget -O models/u2net/u2net.pth https://github.com/xuebinqin/U-2-Net/releases/download/v1.0/u2net.pth

list-models: ## List available models
	@echo "Available models in ./models/:"
	@ls -lh models/

# ============================================
# Docker
# ============================================

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t scrappd-ml:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8000:8000 --env-file .env scrappd-ml:latest

docker-build-gpu: ## Build Docker image with GPU support
	@echo "Building Docker image with GPU support..."
	docker build -t scrappd-ml:latest-gpu --build-arg USE_GPU=true .

docker-run-gpu: ## Run Docker container with GPU
	@echo "Running Docker container with GPU..."
	docker run --gpus all -p 8000:8000 --env-file .env scrappd-ml:latest-gpu

# ============================================
# API Documentation
# ============================================

docs: ## Generate API documentation
	@echo "Starting API documentation server..."
	uvicorn src.api.main:app --host 0.0.0.0 --port 8000
	@echo "Visit http://localhost:8000/docs for Swagger UI"
	@echo "Visit http://localhost:8000/redoc for ReDoc"

# ============================================
# Benchmarking
# ============================================

benchmark: ## Run performance benchmarks
	@echo "Running benchmarks..."
	$(PYTHON) -m pytest tests/benchmark/ -v --benchmark-only

benchmark-memory: ## Run memory profiling
	@echo "Running memory profiling..."
	$(PYTHON) -m memory_profiler scripts/profile_memory.py

profile: ## Profile the application
	@echo "Profiling application..."
	$(PYTHON) -m cProfile -o profile.stats scripts/profile_app.py
	$(PYTHON) -c "import pstats; p = pstats.Stats('profile.stats'); p.sort_stats('cumulative'); p.print_stats(20)"

# ============================================
# Database
# ============================================

redis-cli: ## Connect to Redis
	@echo "Connecting to Redis..."
	redis-cli -h localhost -p $(REDIS_PORT) -a $(REDIS_PASSWORD)

redis-flush: ## Flush Redis cache
	@echo "Flushing Redis cache..."
	redis-cli -h localhost -p $(REDIS_PORT) -a $(REDIS_PASSWORD) FLUSHALL

# ============================================
# Utilities
# ============================================

shell: ## Start Python shell with app context
	@echo "Starting Python shell..."
	$(PYTHON) -i -c "from src.api.main import app; print('App loaded. Use `app` to interact.')"

notebook: ## Start Jupyter notebook
	@echo "Starting Jupyter notebook..."
	jupyter notebook

# ============================================
# Clean
# ============================================

clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".mypy_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	rm -rf htmlcov/
	rm -rf .coverage
	rm -rf dist/
	rm -rf build/
	rm -f profile.stats

clean-models: ## Remove downloaded models
	@echo "Removing models..."
	rm -rf models/*

# ============================================
# Requirements
# ============================================

freeze: ## Freeze current dependencies
	@echo "Freezing dependencies..."
	$(PIP) freeze > requirements.txt

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(PIP) install --upgrade pip
	$(PIP) install --upgrade -r requirements.txt

# ============================================
# Security
# ============================================

security-check: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	safety check
	bandit -r src/

# ============================================
# CI/CD
# ============================================

ci: format-check lint test ## Run CI pipeline
	@echo "CI pipeline complete!"

# ============================================
# Production
# ============================================

prod-install: ## Install production dependencies only
	@echo "Installing production dependencies..."
	$(PIP) install --no-cache-dir -r requirements.txt

# ============================================
# Monitoring
# ============================================

logs: ## Show application logs
	@echo "Showing logs..."
	tail -f logs/app.log

health: ## Check service health
	@echo "Checking service health..."
	curl http://localhost:8000/health