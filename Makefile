.PHONY: help run test test-coverage clean docker-up docker-down

# Variables
APP_NAME=portfolio-rebalancer
MAIN_PATH=./cmd/api

## help: Display available commands
help:
	@echo "Portfolio Rebalancer - Available Commands"
	@echo ""
	@echo "  make run             - Run the application"
	@echo "  make test            - Run all unit tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make clean           - Clean coverage files"
	@echo "  make docker-up       - Start Docker services (Elasticsearch, Kafka)"
	@echo "  make docker-down     - Stop Docker services"
	@echo ""

## run: Run the application
run:
	@echo "Starting application..."
	go run $(MAIN_PATH)/main.go

## test: Run all unit tests
test:
	@echo "Running tests..."
	go test -v ./internal/...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./internal/...
	@go tool cover -func=coverage.out | grep total
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## clean: Clean coverage files
clean:
	@echo "Cleaning..."
	@rm -f coverage.out coverage.html
	@echo "✓ Done"

## docker-up: Start Docker services
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "✓ Services started"

## docker-down: Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "✓ Services stopped"
