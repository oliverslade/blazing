.PHONY: run build clean test generate

run:
	@echo "Starting server..."
	go run cmd/blazing/main.go

build:
	@echo "Building binary..."
	go build -o bin/blazing cmd/blazing/main.go

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

test:
	@echo "Running tests..."
	go test ./...

generate:
	@echo "Generating database code..."
	sqlc generate