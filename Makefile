.PHONY: lint lint-fix test test-coverage build-signaller build-producer build-client clean

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
	@go build -o bin/signaller cmd/signaller/main.go

build-producer:
	@go build -o bin/producer cmd/producer/main.go

build-client:
	@go build -o bin/client cmd/client/main.go

# Clean build artifacts
clean:
	@go clean ./...
	@rm -f coverage.out coverage.html

# Run all checks (lint + test)
check: lint test

generate-swagger:
	@swag init -g cmd/signaller/main.go --output docs --parseDependency --parseInternal --dir .

