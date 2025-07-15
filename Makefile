# Cloud Code Agents - 統合Makefile
# 各サブプロジェクトのMakefileを統合実行

.PHONY: install help clean build test fmt lint install-instructions
.PHONY: hooks-install send-agent-install start-agents-install
.PHONY: hooks-help send-agent-help start-agents-help hooks-setup

# デフォルトターゲット
all: install

# 全プロジェクトのインストール
install: hooks-install send-agent-install start-agents-install install-instructions hooks-setup
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

# Git Hooks セットアップ
hooks-setup:
	@echo "🪝 Setting up Git Hooks..."
	@if [ ! -d ".git" ]; then \
		echo "❌ Error: Not a Git repository"; \
		exit 1; \
	fi
	@mkdir -p .git/hooks
	@# OS検出とシンボリックリンク作成
	@if [ "$$(uname)" = "Darwin" ] || [ "$$(uname)" = "Linux" ]; then \
		if [ -L ".git/hooks/pre-commit" ]; then \
			echo "🔗 Removing existing pre-commit symlink..."; \
			rm -f .git/hooks/pre-commit; \
		elif [ -f ".git/hooks/pre-commit" ]; then \
			echo "📦 Backing up existing pre-commit hook..."; \
			cp .git/hooks/pre-commit .git/hooks/pre-commit.backup; \
			rm -f .git/hooks/pre-commit; \
		fi; \
		echo "🔗 Creating symlink for pre-commit hook..."; \
		ln -sf "$$(pwd)/scripts/hooks/pre-commit" .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
		echo "✅ Git Hooks setup completed with symlinks"; \
	else \
		echo "🪟 Windows detected, using copy method..."; \
		if [ -f ".git/hooks/pre-commit" ]; then \
			echo "📦 Backing up existing pre-commit hook..."; \
			cp .git/hooks/pre-commit .git/hooks/pre-commit.backup; \
		fi; \
		echo "📄 Copying pre-commit hook..."; \
		cp scripts/hooks/pre-commit .git/hooks/pre-commit; \
		chmod +x .git/hooks/pre-commit; \
		echo "✅ Git Hooks setup completed with copy method"; \
	fi
	@echo "🎯 Pre-commit hook installed and ready to use"

# ヘルプの表示
help:
	@echo "🤖 Cloud Code Agents - 統合ビルドシステム"
	@echo ""
	@echo "利用可能なターゲット:"
	@echo "  install           - 全コンポーネントをビルド・インストール"
	@echo "  install-instructions - instructionsフォルダを~/.claude/claude-code-agents/instructionsにコピー"
	@echo "  hooks-setup       - Git Hooksの自動セットアップ"
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
