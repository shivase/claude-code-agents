# Cloud Code Agents - 統合Makefile
# 各サブプロジェクトのMakefileを統合実行

.PHONY: install help clean build test fmt lint install-instructions
.PHONY: hooks-install send-agent-install start-agents-install
.PHONY: hooks-help send-agent-help start-agents-help

# デフォルトターゲット
all: install

# 全プロジェクトのインストール
install: hooks-install send-agent-install start-agents-install install-instructions
	@echo "✅ 全てのコンポーネントのインストールが完了しました"

# 各プロジェクトのインストール
hooks-install:
	@echo "🔧 Installing hooks/reload-role..."
	@$(MAKE) -C hooks/reload-role install

send-agent-install:
	@echo "📨 Installing send-agent..."
	@$(MAKE) -C send-agent install

start-agents-install:
	@echo "🚀 Installing start-agents..."
	@$(MAKE) -C start-agents install

install-instructions:
	@echo "📚 Installing instructions to ~/.claude/claude-code-agents/instructions..."
	@mkdir -p ~/.claude/claude-code-agents/instructions
	@for file in instructions/*; do \
		basename_file=$$(basename "$$file"); \
		target_file="$$HOME/.claude/claude-code-agents/instructions/$$basename_file"; \
		if [ -f "$$target_file" ]; then \
			echo "⚠️  $$basename_file already exists, skipping..."; \
		else \
			cp "$$file" "$$target_file"; \
			echo "✅ Installed $$basename_file"; \
		fi; \
	done
	@echo "✅ Instructions installation completed"

# ヘルプの表示
help:
	@echo "🤖 Cloud Code Agents - 統合ビルドシステム"
	@echo ""
	@echo "利用可能なターゲット:"
	@echo "  install           - 全コンポーネントをビルド・インストール"
	@echo "  install-instructions - instructionsフォルダを~/.claude/claude-code-agents/instructionsにコピー"
	@echo "  help              - このヘルプメッセージを表示"
	@echo "  clean             - 全プロジェクトのビルド成果物をクリーンアップ"
	@echo "  build             - 全プロジェクトをビルド"
	@echo "  test              - 全プロジェクトのテストを実行"
	@echo "  fmt               - 全プロジェクトのコードをフォーマット"
	@echo "  lint              - 全プロジェクトのリント実行"
	@echo ""
	@echo "各プロジェクトの詳細については個別のヘルプを参照してください。"

# 全プロジェクトのクリーンアップ
clean:
	@echo "🧹 Cleaning all projects..."
	@$(MAKE) -C hooks/reload-role clean
	@$(MAKE) -C send-agent clean
	@$(MAKE) -C start-agents clean
	@echo "✅ All projects cleaned"

# 全プロジェクトのビルド
build:
	@echo "🔨 Building all projects..."
	@$(MAKE) -C hooks/reload-role build
	@$(MAKE) -C send-agent build
	@$(MAKE) -C start-agents build
	@echo "✅ All projects built"

# 全プロジェクトのテスト
test:
	@echo "🧪 Testing all projects..."
	@$(MAKE) -C hooks/reload-role test
	@$(MAKE) -C send-agent test
	@$(MAKE) -C start-agents test
	@echo "✅ All tests completed"

# 全プロジェクトのフォーマット
fmt:
	@echo "🎨 Formatting all projects..."
	@$(MAKE) -C hooks/reload-role fmt
	@$(MAKE) -C send-agent fmt
	@$(MAKE) -C start-agents fmt
	@echo "✅ All projects formatted"

# 全プロジェクトのリント
lint:
	@echo "🔍 Linting all projects..."
	@$(MAKE) -C hooks/reload-role lint
	@$(MAKE) -C send-agent lint
	@$(MAKE) -C start-agents lint
	@echo "✅ All projects linted"
