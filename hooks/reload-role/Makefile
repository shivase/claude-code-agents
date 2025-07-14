# Makefile for reload-role Go application

# Variables
GO_VERSION := 1.21
BINARY_NAME := reload-role
GO_FILES := $(wildcard *.go)
BUILD_DIR := build
INSTALL_DIR := ~/.claude/hooks

# マルチプラットフォーム対応
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# ビルドフラグ
LDFLAGS := -s -w

# Default target
.PHONY: all
all: build

# 依存関係の取得
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Build the application
.PHONY: build
build: deps $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES) go.mod
	@echo "🔨 Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

# クロスコンパイル用ターゲット
.PHONY: build-linux build-darwin build-windows build-all
build-linux:
	@echo "🔨 Building $(BINARY_NAME) for linux/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

build-darwin:
	@echo "🔨 Building $(BINARY_NAME) for darwin/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

build-windows:
	@echo "🔨 Building $(BINARY_NAME) for windows/amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

build-all: build-linux build-darwin build-windows
	@echo "✅ All platform builds completed"

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Install the binary to system PATH
install: build
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

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application for current platform"
	@echo "  build-linux  - Build for Linux (amd64)"
	@echo "  build-darwin - Build for macOS (amd64)"
	@echo "  build-windows - Build for Windows (amd64)"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  install      - Install binary to $(INSTALL_DIR)"
	@echo "  uninstall    - Remove binary from $(INSTALL_DIR)"
	@echo "  test         - Run tests"
	@echo "  run          - Build and run the application"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Run linters"
	@echo "  mod-update   - Update Go modules"
	@echo "  help         - Show this help message"

.PHONY: all deps build clean install uninstall test run fmt lint mod-update help
