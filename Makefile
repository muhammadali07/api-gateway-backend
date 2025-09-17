.PHONY: build run test clean docker-up docker-down docker-logs help

# Variables
APP_NAME=api-gateway-backend
DOCKER_COMPOSE=docker-compose
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*')

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	go run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short: ## Run tests without coverage
	@echo "Running tests..."
	go test -v ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w $(GO_FILES)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

docker-up: ## Start services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started. API available at http://localhost:8080"
	@echo "Use 'make docker-logs' to view logs"

docker-down: ## Stop services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker Compose logs
	$(DOCKER_COMPOSE) logs -f

docker-restart: docker-down docker-up ## Restart services

# Database
db-migrate: ## Run database migrations (placeholder)
	@echo "Database migrations would run here"
	@echo "Currently using init.sql in docker-compose"

db-seed: ## Seed database with test data (placeholder)
	@echo "Database seeding would run here"
	@echo "Currently using sample data in init.sql"

# Development helpers
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	go mod download
	go mod tidy
	@echo "Development environment ready!"

dev-reset: docker-down clean docker-up ## Reset development environment

# API testing
api-test: ## Test API endpoints (requires running service)
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -s http://localhost:8080/health | jq .
	@echo "\nSync data:"
	curl -s -X POST http://localhost:8080/api/v1/sync | jq .
	@echo "\nGet items:"
	curl -s http://localhost:8080/api/v1/items | jq .
	@echo "\nOrder status summary:"
	curl -s http://localhost:8080/api/v1/analytics/orders/status | jq .
	@echo "\nTop customers:"
	curl -s http://localhost:8080/api/v1/analytics/customers/top | jq .

# Monitoring
monitor: ## Show service status
	@echo "Service Status:"
	$(DOCKER_COMPOSE) ps
	@echo "\nResource Usage:"
	docker stats --no-stream