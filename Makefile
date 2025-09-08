# Makefile for git-assist

# Variables
BINARY_NAME=git-assist
VERSION=0.1.0
BUILD_DIR=build
MAIN_PATH=cmd/git-assist/main.go
COVERAGE_FILE=coverage.out

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# Default target
.PHONY: all
all: clean deps fmt test build

# Build the binary
.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Development build with debug info and race detection
.PHONY: dev-build
dev-build:
	$(GOBUILD) -race -gcflags="all=-N -l" $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-test
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE)
	rm -f coverage.html

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run unit tests
.PHONY: test
test:
	$(GOTEST) -v ./internal/...

# Run unit tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) ./internal/...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests
.PHONY: test-integration
test-integration: build
	$(GOTEST) -v ./tests/integration/...

# Run all tests
.PHONY: test-all
test-all: test test-integration

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GOTEST) -race -v ./internal/...

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Vet code
.PHONY: vet
vet:
	$(GOCMD) vet ./...

# Cross-compile for different platforms
.PHONY: build-all
build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Create release packages
.PHONY: package
package: build-all
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip $(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

# Install locally
.PHONY: install
install: build
	@if [ -d "$$HOME/bin" ]; then \
		cp $(BINARY_NAME) $$HOME/bin/; \
		echo "Installed to $$HOME/bin/$(BINARY_NAME)"; \
	elif [ -w "/usr/local/bin" ]; then \
		cp $(BINARY_NAME) /usr/local/bin/; \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "Please copy $(BINARY_NAME) to a directory in your PATH"; \
	fi

# Pre-commit checks
.PHONY: pre-commit
pre-commit: fmt vet test-all

# Release preparation
.PHONY: release-prep
release-prep: clean fmt vet test-all build-all package
	@echo "Release $(VERSION) prepared in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Help
.PHONY: help
help:
	@echo "git-assist Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  dev-build      - Development build with debug info"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run unit tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all       - Run all tests"
	@echo "  test-race      - Run tests with race detection"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo ""
	@echo "Build & Release:"
	@echo "  build-all      - Cross-compile for all platforms"
	@echo "  package        - Create release packages"
	@echo "  release-prep   - Prepare release"
	@echo ""
	@echo "Installation:"
	@echo "  install        - Install locally"
	@echo ""
	@echo "Utilities:"
	@echo "  pre-commit     - Pre-commit checks"
	@echo "  help           - Show this help"
