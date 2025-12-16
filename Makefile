.PHONY: help build run test test-container inttest docker-build docker-build-test docker-run swagger clean deps install-swag

# Default target
.DEFAULT_GOAL := help

# Show help message
help:
	@echo "Bitcoin LTP API - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  help              Show this help message"
	@echo "  build             Build the application"
	@echo "  run               Run the application"
	@echo "  test              Run all unit tests"
	@echo "  inttest           Run integration tests (requires Docker, builds image first)"
	@echo "  docker-build-test Build Docker image for testing"
	@echo "  docker-build      Build Docker image"
	@echo "  docker-run        Run Docker container"
	@echo "  swagger           Generate Swagger documentation"
	@echo "  clean             Clean build artifacts"
	@echo "  deps              Install/update dependencies"
	@echo "  install-swag      Install swag tool for Swagger"
	@echo ""

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run all tests
test:
	go test -v ./...

# Run container tests
test-container:
	go test -v ./tests/integration/...

# Run integration tests
inttest: docker-build-test
	go test -v ./tests/integration/...

# Build Docker image for testing
docker-build-test:
	docker build -t bitcoin-ltp-api:test .

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs

# Build Docker image
docker-build:
	docker build -t bitcoin-ltp-api:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 bitcoin-ltp-api:latest

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install swag tool
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

