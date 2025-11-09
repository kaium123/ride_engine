.PHONY: help build run docker-up docker-down docker-logs test clean swagger-init swagger-update swagger

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/ride_engine main.go
	@echo "Build complete: bin/ride_engine"

run: ## Run the application
	@echo "Running application..."
	@go run main.go serve

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Waiting for databases to be ready..."
	@sleep 5
	@echo "Docker containers are up!"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-restart: ## Restart Docker containers
	@echo "Restarting Docker containers..."
	@docker-compose restart

docker-logs: ## Show Docker container logs
	@docker-compose logs -f

docker-clean: ## Remove Docker containers and volumes
	@echo "Removing Docker containers and volumes..."
	@docker-compose down -v
	@echo "Cleanup complete!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded!"

migrate-up: ## Run database migrations (placeholder)
	@echo "Running migrations..."
	@echo "Note: Migrations run automatically via Docker init scripts"

db-reset: docker-down docker-clean docker-up ## Reset databases
	@echo "Databases reset complete!"

swagger-init: ## Initialize Swagger documentation (first time setup)
	@echo "Installing swag CLI tool..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Initializing Swagger documentation..."
	@swag init -g main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation initialized in docs/ directory"
	@echo "View at: http://localhost:8080/swagger/index.html"

swagger-update: ## Update Swagger documentation
	@echo "Updating Swagger documentation..."
	@swag init -g main.go -o docs --parseDependency --parseInternal
	@echo "Swagger documentation updated!"
	@echo "View at: http://localhost:8080/swagger/index.html"

swagger: swagger-update ## Alias for swagger-update

.DEFAULT_GOAL := help
