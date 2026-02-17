.PHONY: build run test lint swagger docker-build docker-up docker-down migrate seed clean help

# Variables
BINARY_NAME=threat-intel-api
GO=go

# Help
help:
	@echo "Threat Intelligence API - Available commands:"
	@echo ""
	@echo "  make build        - Build the application binary"
	@echo "  make run          - Run the application locally"
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage report"
	@echo "  make lint         - Run linter"
	@echo "  make seed         - Seed the database with sample data"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start all services with Docker Compose"
	@echo "  make docker-down  - Stop all Docker services"
	@echo "  make docker-logs  - View Docker logs"
	@echo "  make clean        - Remove build artifacts"
	@echo ""

# Build
build:
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/api

# Run locally
run:
	$(GO) run ./cmd/api

# Tests
test:
	$(GO) test -v -race ./...

test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Seed database
seed:
	$(GO) run ./scripts/seed.go

# Docker commands
docker-build:
	docker build -t $(BINARY_NAME):latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart api

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development - run with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Format code
fmt:
	$(GO) fmt ./...

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy
