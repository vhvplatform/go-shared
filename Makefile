.PHONY: help test test-coverage lint fmt vet tidy clean build install

# Variables
GO := go
GOFLAGS := -v
GOTEST := $(GO) test
GOLINT := golangci-lint
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Go version check
GO_VERSION := 1.25.5
CURRENT_GO_VERSION := $(shell $(GO) version | awk '{print $$3}' | sed 's/go//')

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@echo ""
	@echo "  make test              - Run all tests"
	@echo "  make test-fast         - Run tests without race detector (faster)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make bench             - Run benchmark tests"
	@echo "  make lint              - Run linters"
	@echo "  make fmt               - Format code"
	@echo "  make vet               - Run go vet"
	@echo "  make tidy              - Tidy and verify dependencies"
	@echo "  make clean             - Clean build artifacts and cache"
	@echo "  make build             - Build all packages"
	@echo "  make build-fast        - Build with optimizations (faster)"
	@echo "  make install           - Install dependencies"
	@echo "  make check             - Run all checks (fmt, vet, lint, test)"
	@echo "  make ci                - Run CI pipeline locally"
	@echo ""

## install: Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod verify

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## lint: Run golangci-lint
lint:
	@echo "Running linters..."
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		export PATH="$$PATH:$$(go env GOPATH)/bin"; \
	fi
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && golangci-lint run ./... --timeout=5m

## build: Build all packages
build:
	@echo "Building all packages..."
	$(GO) build $(GOFLAGS) ./...

## build-fast: Build all packages with caching (faster builds)
build-fast:
	@echo "Building all packages with optimizations..."
	GOCACHE=$$(go env GOCACHE) $(GO) build -trimpath $(GOFLAGS) ./...

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -race -v ./...

## test-fast: Run tests without race detector (faster)
test-fast:
	@echo "Running tests (fast mode)..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Generating coverage report..."
	$(GO) tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "Coverage summary:"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'
	@echo ""
	@echo "To view HTML coverage report, run: make coverage-html"

## coverage-html: Generate HTML coverage report
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@echo "Open it in your browser to view"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"

## ci: Run CI pipeline locally
ci: install check test-coverage
	@echo "CI pipeline completed successfully!"

## clean: Clean build artifacts and cache
clean:
	@echo "Cleaning..."
	$(GO) clean -cache -testcache -modcache
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete!"

## version: Display Go version
version:
	@echo "Current Go version: $(CURRENT_GO_VERSION)"
	@echo "Required Go version: $(GO_VERSION)"
	@if [ "$(CURRENT_GO_VERSION)" != "$(GO_VERSION)" ]; then \
		echo "WARNING: Go version mismatch! Please use Go $(GO_VERSION)"; \
	fi

## bench: Run benchmark tests
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem -run=^$$ ./...
