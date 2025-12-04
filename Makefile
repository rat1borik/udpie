.PHONY: lint lint-fix test test-coverage build clean

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
build:
	@go build ./...

# Clean build artifacts
clean:
	@go clean ./...
	@rm -f coverage.out coverage.html

# Run all checks (lint + test)
check: lint test

