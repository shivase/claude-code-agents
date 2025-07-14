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
	sudo rm $(INSTALL_DIR)/$(BINARY_NAME)
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
	go test -p 1 -v -timeout=10m -count=1 ./test/...

# カバレッジ付きテスト実行
.PHONY: test-coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./test/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📄 Coverage report generated: coverage.html"

# カバレッジ閾値チェック
.PHONY: test-coverage-check
test-coverage-check:
	@echo "🎯 Checking coverage thresholds..."
	go test -coverprofile=coverage.out ./test/...
	@go tool cover -func=coverage.out | grep "total:" | awk '{if($$3+0 < 75.0) {print "❌ Coverage below 75%: " $$3; exit 1} else {print "✅ Coverage above 75%: " $$3}}'

.PHONY: lint
lint:
	@echo "🔍 Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run internal/... --config .golangci.yaml; \
		golangci-lint run test/... --config .golangci.yaml; \
	else \
		echo "⚠️  golangci-lint not found, running go vet instead"; \
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
	@echo "  test               - Run all tests"
	@echo "  test-ci            - Run tests in CI environment with mocks"
	@echo "  test-coverage      - Run tests with coverage and generate HTML report"
	@echo "  test-coverage-check- Check coverage meets 75% threshold"
	@echo "  lint               - Run static analysis"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts and caches"
	@echo "  help               - Show this help message"
