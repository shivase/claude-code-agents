# Cloud Code Agents

tmux上でClaude Code AIエージェントを並列実行する統合開発環境システムです。

## 概要

このプロジェクトは、複数のAIエージェントを並列で実行し、チーム開発を効率化するためのツールセットです。以下の2つの主要コンポーネントで構成されています：

- **start-agents**: AIエージェントセッションを起動・管理するメインシステム
- **send-agent**: 起動中のエージェントにメッセージを送信するクライアントツール

## 機能

### 🚀 start-agents (AIエージェントセッション起動システム)

複数のClaude Code AIエージェントを並列で起動し、管理します。

**主な機能：**
- 統合監視画面での6ペイン並列実行
- 個別セッション方式での分散実行
- セッション管理（作成、削除、一覧表示）
- 設定ファイルによる柔軟な環境設定
- 詳細ログ・サイレントモード対応

**利用可能なエージェント：**
- `ceo`: 最高経営責任者（全体統括）
- `manager`: プロジェクトマネージャー（チーム管理）
- `dev1-dev4`: 実行エージェント（柔軟な役割対応）

### 📤 send-agent (メッセージ送信システム)

起動中のAIエージェントにメッセージを送信します。

**主な機能：**
- エージェント別メッセージ送信
- 統合監視画面・個別セッション両方に対応
- 自動セッション検出
- 通信ログの自動記録


## 設定

### 設定ファイル生成

```bash
# 設定ファイルのテンプレート生成
claude-code-agents --generate-config

# 既存ファイルを上書きして生成
claude-code-agents --generate-config --force
```

### 環境変数

```bash
# 詳細ログ有効化
export VERBOSE=true

# サイレントモード有効化
export SILENT=true
```

## 技術仕様

### システム要件

- Go 1.21以上
- tmux
- Claude Code CLI

### アーキテクチャ

- **start-agents**: Go言語で実装されたメインシステム
- **send-agent**: Go言語で実装されたクライアントツール
- **tmux**: セッション管理とペイン分割
- **Claude Code CLI**: AIエージェントエンジン

## 開発

### ビルド

```bash
# start-agents
cd start-agents
make build

# send-agent
cd send-agent
make build
```

### テスト

```bash
# start-agents
cd start-agents
make test

# send-agent
cd send-agent
make test
```

## ライセンス

このプロジェクトは適切なライセンスの下で提供されます。

## 貢献

プロジェクトへの貢献を歓迎します。Issue報告やPull Requestをお待ちしています。
