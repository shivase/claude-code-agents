# Makefile for send-agent Go application

# Variables
GO_VERSION := 1.21
BINARY_NAME := send-agent
GO_FILES := $(wildcard *.go)
BUILD_DIR := build
INSTALL_DIR := /usr/local/bin
PACKAGE_NAME := send-agent

# マルチプラットフォーム対応
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# ビルドフラグ
LDFLAGS := -s -w

# Default target
all: fmt install

# Build the application
.PHONY: build
build: $(GO_FILES) go.mod
	@echo "🔨 Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="$(LDFLAGS)" .
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Install the binary to system PATH
install: build
	sudo rm $(INSTALL_DIR)/$(BINARY_NAME)
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

# Uninstall the binary from system PATH
uninstall:
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)

# Test the application
test:
	go test -v ./...

# Run the application
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Format Go code
fmt:
	go fmt ./...


.PHONY: lint
lint:
	@echo "🔍 Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run *.go --config .golangci.yaml; \
	else \
		echo "⚠️  golangci-lint not found, running go vet instead"; \
		go vet ./...; \
	fi

# Check for Go modules updates
mod-update:
	go mod tidy
	go mod download

# Development build (with race detector)
.PHONY: dev-build
dev-build:
	@echo "🔨 Building $(BINARY_NAME) for $(GOOS)/$(GOARCH) with race detector..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -race -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Development build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# マルチプラットフォームビルド
.PHONY: build-all
build-all:
	@echo "🔨 Building $(BINARY_NAME) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ] && [ "$$arch" = "arm64" ]; then \
				continue; \
			fi; \
			echo "Building for $$os/$$arch..."; \
			BINARY_PATH="$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch"; \
			if [ "$$os" = "windows" ]; then \
				BINARY_PATH="$$BINARY_PATH.exe"; \
			fi; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o "$$BINARY_PATH" -ldflags="$(LDFLAGS)" .; \
		done; \
	done
	@echo "✅ Multi-platform build completed"

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application for current platform ($(GOOS)/$(GOARCH))"
	@echo "  build-all   - Build for multiple platforms (linux, darwin, windows)"
	@echo "  dev-build   - Build with race detector for development"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install binary to $(INSTALL_DIR)"
	@echo "  uninstall   - Remove binary from $(INSTALL_DIR)"
	@echo "  test        - Run tests"
	@echo "  run         - Build and run the application"
	@echo "  fmt         - Format Go code"
	@echo "  lint        - Run linters"
	@echo "  mod-update  - Update Go modules"
	@echo "  help        - Show this help message"

.PHONY: all build build-all dev-build clean install uninstall test run fmt lint mod-update help
