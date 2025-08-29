# Makefile for Pub/Sub System

# Binary name
BINARY_NAME=pub-sub

# Build the application
build:
	go build -o $(BINARY_NAME) .

# Run the application
run: build
	./$(BINARY_NAME)

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Install dependencies
deps:
	go mod tidy
	go mod download

# Setup environment
env-setup:
	cp .env.example .env
	@echo "Environment file created. Edit .env with your configuration values."

# Build for Docker
docker-build:
	docker build -t pub-sub .

# Run with Docker
docker-run:
	docker run -p 8080:8080 pub-sub

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Test logging system
test-logging: build
	@echo "Testing logging system with different levels and formats..."
	@echo "=== Text format with INFO level ==="
	LOG_LEVEL=info LOG_FORMAT=text timeout 3s ./$(BINARY_NAME) || true
	@echo ""
	@echo "=== JSON format with DEBUG level ==="
	LOG_LEVEL=debug LOG_FORMAT=json timeout 3s ./$(BINARY_NAME) || true
	@echo ""
	@echo "Logging test completed!"

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  env-setup      - Setup environment file from template"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  test-logging   - Test logging system with different formats"
	@echo "  help           - Show this help message"

.PHONY: build run test test-coverage clean deps env-setup docker-build docker-run fmt lint test-logging help
