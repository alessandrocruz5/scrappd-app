.PHONY: help setup install run test clean docker-build docker-run docker-stop

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

setup: ## Create virtual environment
	python3 -m venv venv
	@echo "✓ Virtual environment created"
	@echo "Run 'source venv/bin/activate' to activate it"

install: ## Install dependencies
	pip install --upgrade pip
	pip install -r requirements.txt
	@echo "✓ Dependencies installed"

dev: ## Run development server with auto-reload
	uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

run: ## Run production server
	uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 2

test: ## Run test script
	python test_service.py

test-image: ## Test with specific image (usage: make test-image IMG=path/to/image.jpg)
	python test_service.py $(IMG)

clean: ## Clean up generated files
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete
	find . -type f -name "test_output.png" -delete
	@echo "✓ Cleaned up"

clean-all: clean ## Clean up everything including venv
	rm -rf venv
	@echo "✓ Cleaned everything"

docker-build: ## Build Docker image
	docker build -t scrappd-ml-service:latest .

docker-run: ## Run Docker container
	docker run -d --name scrappd-ml -p 8000:8000 scrappd-ml-service:latest

docker-stop: ## Stop and remove Docker container
	docker stop scrappd-ml || true
	docker rm scrappd-ml || true

docker-logs: ## Show Docker container logs
	docker logs -f scrappd-ml

docker-shell: ## Open shell in running container
	docker exec -it scrappd-ml /bin/bash

docker-test: ## Test Dockerized service
	@echo "Waiting for service to start..."
	@sleep 3
	curl http://localhost:8000/health

info: ## Show service information
	@echo "Service: Scrappd ML Service"
	@echo "Port: 8000"
	@echo "Health: http://localhost:8000/health"
	@echo "Process: http://localhost:8000/process"