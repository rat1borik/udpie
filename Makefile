.PHONY: lint lint-fix test test-coverage build build-signaller build-producer build-client build-consumer clean help

# Install golangci-lint if not present
lint-install:
	@echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
lint:
	@golangci-lint run -v --fix ./...

# Run tests
test:
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build all binaries
build-signaller:
	@go build -o build/signaller/signaller.exe ./cmd/signaller

build-producer:
	@go build -o build/producer/producer.exe ./cmd/producer

build-consumer:
	@go build -o build/consumer/consumer.exe ./cmd/consumer

# Build all binaries
build: build-signaller build-producer build-consumer
	@echo "All binaries built successfully"

# Clean build artifacts
clean:
	@go clean ./...
	@rm -f coverage.out coverage.html

# Run all checks (lint + test)
check: lint test

generate-swagger:
	@swag init -g cmd/signaller/main.go --output docs --parseDependency --parseInternal --dir .

# Show help message
help:
	@echo "Available targets:"
	@echo "  build            - Build all binaries (signaller, producer, consumer)"
	@echo "  build-signaller  - Build signaller binary"
	@echo "  build-producer   - Build producer binary"
	@echo "  build-consumer   - Build consumer binary"
	@echo "  lint             - Run linters"
	@echo "  lint-install     - Install golangci-lint"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  check            - Run lint and test"
	@echo "  clean            - Clean build artifacts"
	@echo "  generate-swagger - Generate Swagger documentation"

