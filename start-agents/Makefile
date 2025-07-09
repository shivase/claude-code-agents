# Go言語版Claude CLI管理システム

# 変数定義
GO_VERSION := 1.21
BINARY_NAME := claude-code-agents
BUILD_DIR := build
INSTALL_DIR=/usr/local/bin
PACKAGE_NAME := claude-code-agents

# マルチプラットフォーム対応
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# ビルドフラグ
LDFLAGS := -s -w

# デフォルトターゲット
.PHONY: all
all: build

# 依存関係の取得
.PHONY: deps
deps:
	go mod tidy
	go mod download

# ビルド
.PHONY: build
build: deps
	@echo "🔨 Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="$(LDFLAGS)" *.go
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# インストール
.PHONY: install
install: build
	@echo "📦 Installing binaries to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

# アンインストール
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling binaries..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Uninstallation completed"

# テスト
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

.PHONY: lint
lint:
	@echo "🔍 Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		@echo "⚠️  golangci-lint not found, running go vet instead"; \
		go vet ./...; \
	fi

# フォーマット
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# クリーンアップ
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean -cache
	go clean -modcache
	@echo "✅ Cleanup completed"

# ヘルプ
.PHONY: help
help:
	@echo "🤖 Claude Code Agents - Go Implementation"
	@echo ""
	@echo "Available targets:"
	@echo "  build              - Build the main binary for current platform"
	@echo "  install            - Install binaries to $(INSTALL_DIR)"
	@echo "  uninstall          - Remove installed binaries"
	@echo "  test               - Run tests"
	@echo "  lint               - Run static analysis"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts and caches"
	@echo "  help               - Show this help message"
