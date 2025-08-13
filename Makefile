# Include .env file if it exists
-include .env
export

# Default target
.PHONY: all
all: clean test build

## Build the binary
.PHONY: build
build:
	@echo "Building blazing..."
	@mkdir -p bin
	go build -ldflags="-w -s" -o bin/blazing ./cmd/server

## Run the application in development mode
.PHONY: dev
dev: generate
	@echo "Starting development server..."
	go run ./cmd/server/main.go

## Run the built binary
.PHONY: run
run: build
	@echo "Running blazing..."
	./bin/blazing

## Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

## Run tests with coverage
.PHONY: coverage
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run all code quality checks
.PHONY: check
check:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Running go vet..."
	go vet ./...
	@echo "Running tests..."
	go test ./...
	@echo "All checks passed!"

## Generate database code
.PHONY: generate
generate:
	@echo "Generating database code..."
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	else \
		echo "sqlc not installed. Install with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
	fi

## Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf bin/ coverage.out coverage.html

## Clean database files
.PHONY: clean-db
clean-db:
	@echo "Cleaning database files..."
	@rm -f *.db *.db-shm *.db-wal

## Clean everything
.PHONY: clean-all
clean-all: clean clean-db
