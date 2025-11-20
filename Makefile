# TrainBlink Server Makefile

.PHONY: help run build test clean fmt lint deps docker-up docker-down

# Variables
BINARY_NAME=trainblink-server
GO=go
GOFLAGS=-v
BUILD_DIR=bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

run: ## Run the server
	$(GO) run cmd/server/main.go

build: ## Build the server binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage: test ## Generate test coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean

docker-up: ## Start Docker services (PostgreSQL, Redis)
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

migrate-up: ## Run database migrations
	@echo "Database migrations not yet implemented (Phase 2)"

migrate-down: ## Rollback database migrations
	@echo "Database migrations not yet implemented (Phase 2)"

.PHONY: phase0-test
phase0-test: ## Test Phase 0 endpoints
	@echo "Testing Phase 0 endpoints..."
	@echo ""
	@echo "1. Testing /ping..."
	@curl -s http://localhost:8080/ping | jq '.'
	@echo ""
	@echo "2. Testing /health..."
	@curl -s http://localhost:8080/health | jq '.'
	@echo ""
	@echo "3. Testing /api/v1/hello..."
	@curl -s http://localhost:8080/api/v1/hello | jq '.'
	@echo ""
	@echo "4. Testing /api/v1/welcome..."
	@curl -s "http://localhost:8080/api/v1/welcome?name=Alice&device_id=iOS-1234" | jq '.'
	@echo ""
	@echo "All tests complete!"
