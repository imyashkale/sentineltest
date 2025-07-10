# WafGuard Makefile

# Variables
BINARY_NAME=wafguard
MAIN_PATH=./cmd/wafguard
BUILD_DIR=./bin
COVERAGE_DIR=./coverage
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

.PHONY: all build clean test test-coverage test-race help install install-local uninstall deps lint format vet check dev-deps examples

# Default target
all: clean deps lint test build

# Help target
help: ## Show this help message
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BINARY_NAME) in $(BUILD_DIR)/"

# Build for multiple platforms
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Built binaries for multiple platforms in $(BUILD_DIR)/"

# Install the application globally
install: build ## Install the application to /usr/local/bin (requires sudo)
	@echo "Installing $(BINARY_NAME) globally to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		echo "$(BINARY_NAME) installed successfully to /usr/local/bin/"; \
	else \
		echo "Installing $(BINARY_NAME) to /usr/local/bin (requires sudo)..."; \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
		echo "$(BINARY_NAME) installed successfully to /usr/local/bin/"; \
	fi
	@echo "You can now run '$(BINARY_NAME)' from anywhere!"

# Install to local Go bin (no sudo required)
install-local: ## Install the application to GOPATH/bin or ~/go/bin
	@echo "Installing $(BINARY_NAME) locally..."
	@if [ -n "$(GOPATH)" ]; then \
		$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH); \
		echo "$(BINARY_NAME) installed to $(GOPATH)/bin/"; \
	else \
		$(GOBUILD) $(LDFLAGS) -o $(HOME)/go/bin/$(BINARY_NAME) $(MAIN_PATH); \
		echo "$(BINARY_NAME) installed to $(HOME)/go/bin/"; \
	fi
	@echo "Make sure your Go bin directory is in your PATH:"
	@echo "  export PATH=\$$PATH:\$$GOPATH/bin  # or"
	@echo "  export PATH=\$$PATH:\$$HOME/go/bin"

# Uninstall the application
uninstall: ## Uninstall the application from system
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ -f /usr/local/bin/$(BINARY_NAME) ]; then \
		if [ -w /usr/local/bin ]; then \
			rm -f /usr/local/bin/$(BINARY_NAME); \
		else \
			sudo rm -f /usr/local/bin/$(BINARY_NAME); \
		fi; \
		echo "$(BINARY_NAME) removed from /usr/local/bin/"; \
	fi
	@if [ -n "$(GOPATH)" ] && [ -f $(GOPATH)/bin/$(BINARY_NAME) ]; then \
		rm -f $(GOPATH)/bin/$(BINARY_NAME); \
		echo "$(BINARY_NAME) removed from $(GOPATH)/bin/"; \
	fi
	@if [ -f $(HOME)/go/bin/$(BINARY_NAME) ]; then \
		rm -f $(HOME)/go/bin/$(BINARY_NAME); \
		echo "$(BINARY_NAME) removed from $(HOME)/go/bin/"; \
	fi
	@echo "$(BINARY_NAME) has been uninstalled."

# Clean build artifacts
clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned build artifacts"

# Download dependencies
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o $(COVERAGE_DIR)/coverage.html
	$(GOCMD) tool cover -func=coverage.out
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Run tests with race detection
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -v -race ./...

# Benchmark tests
benchmark: ## Run benchmark tests
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Lint the code
lint: ## Run linter
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not installed. Install with: make dev-deps"; \
	fi

# Format the code
format: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Code formatted"

# Vet the code
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Check code quality
check: format vet lint ## Run all code quality checks

# Install development dependencies
dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@echo "Development dependencies installed"

# Run the application with example config
run-example: build ## Run the application with example configuration
	@echo "Running example..."
	$(BUILD_DIR)/$(BINARY_NAME) run examples/test-configs/sql-injection-test.yaml

# Validate example configurations
validate-examples: build ## Validate example configurations
	@echo "Validating example configurations..."
	$(BUILD_DIR)/$(BINARY_NAME) validate examples/test-configs/

# Create example test cases
examples: ## Create and run example test cases
	@echo "Running example test cases..."
	@echo "1. Validating configurations..."
	$(BUILD_DIR)/$(BINARY_NAME) validate examples/test-configs/ || true
	@echo "2. Running SQL injection tests..."
	$(BUILD_DIR)/$(BINARY_NAME) run examples/test-configs/sql-injection-test.yaml --format json || true
	@echo "3. Running XSS tests..."
	$(BUILD_DIR)/$(BINARY_NAME) run examples/test-configs/xss-test.yaml --concurrent 3 || true

# Development workflow
dev: clean deps check test build ## Complete development workflow

# Release workflow
release: clean deps check test test-race build-all ## Complete release workflow

# Docker build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker build -t $(BINARY_NAME):latest .

# Docker run
docker-run: docker-build ## Run application in Docker
	docker run --rm -v $(PWD)/examples:/app/examples $(BINARY_NAME):latest validate /app/examples/test-configs/

# Security scan
security: ## Run security scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@mkdir -p docs
	$(GOCMD) doc -all . > docs/api.md
	@echo "Documentation generated in docs/"

# Show project statistics
stats: ## Show project statistics
	@echo "Project Statistics:"
	@echo "=================="
	@echo "Go files: $$(find . -name '*.go' | wc -l)"
	@echo "Lines of code: $$(find . -name '*.go' -exec cat {} \; | wc -l)"
	@echo "Test files: $$(find . -name '*_test.go' | wc -l)"
	@echo "Packages: $$(go list ./... | wc -l)"

# Continuous Integration target
ci: deps check test-race test-coverage ## CI pipeline target

# Watch for changes and run tests
watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make test; \
	else \
		echo "entr not installed. Install with your package manager"; \
		echo "macOS: brew install entr"; \
		echo "Ubuntu: apt-get install entr"; \
	fi

# Update dependencies
update-deps: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Version information
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build time: $$(date)"

# Pre-commit hook setup
setup-hooks: ## Setup pre-commit hooks
	@echo "Setting up pre-commit hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake check test' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hooks installed"