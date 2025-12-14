# InterviewMin Backend Makefile
# ============================================================================

# Build variables
BINARY_NAME := app
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Linker flags for injecting version info
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)

# Go build parameters
GO := go
GOFLAGS := -tags netgo

.PHONY: all build build-local run clean test lint help

# Default target
all: build

## build: Build production-ready binary (for deployment)
build:
	@echo "Building $(BINARY_NAME) $(VERSION) ($(COMMIT))..."
	$(GO) mod tidy
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/api
	@echo "Build complete: ./$(BINARY_NAME)"

## build-local: Build binary for local development
build-local:
	@echo "Building $(BINARY_NAME) for local development..."
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/api
	@echo "Build complete: ./$(BINARY_NAME)"

## run: Run the application locally (uses air for hot reload if available)
run:
	@if command -v air > /dev/null; then \
		air; \
	else \
		$(GO) run ./cmd/api; \
	fi

## run-binary: Run the compiled binary
run-binary: build-local
	./$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf tmp/
	@echo "Clean complete"

## test: Run all tests
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

## test-short: Run short tests only
test-short:
	$(GO) test -v -short ./...

## lint: Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## vet: Run go vet
vet:
	$(GO) vet ./...

## fmt: Format code
fmt:
	$(GO) fmt ./...

## tidy: Clean up go.mod and go.sum
tidy:
	$(GO) mod tidy

## version: Display version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
