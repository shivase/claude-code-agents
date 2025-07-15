# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 開発コマンド

### ビルドとテスト
```bash
# 依存関係の取得とビルド
make build

# テストの実行
make test

# リンターとフォーマットの実行
make lint
make fmt

# 単一テストの実行
go test -v ./internal/config -run TestLoadConfig
```

### インストールとクリーンアップ
```bash
# システムへのインストール
make install

# 成果物のクリーンアップ
make clean
```

### セッション管理
```bash
# セッション作成（統合監視画面）
./build/claude-code-agents ai-teams

# セッション作成（個別セッション）
./build/claude-code-agents ai-teams --layout individual

# 既存セッションのリセット
./build/claude-code-agents ai-teams --reset

# セッション一覧表示
./build/claude-code-agents --list

# セッション削除
./build/claude-code-agents --delete ai-teams

# 設定表示
./build/claude-code-agents --show-config
```

## アーキテクチャ概要

### 主要コンポーネント

1. **cmd**: コマンドライン引数の解析とサブコマンド処理
   - 起動オプションの管理（--debug, --verbose, --reset等）
   - セッション操作コマンド（--list, --delete等）
   - 設定表示コマンド（--show-config, --config等）

2. **config**: アプリケーション設定とリソース監視
   - JSON設定ファイルの読み込み/保存
   - デフォルト設定の管理（~/.claude/claude-code-agents/）
   - リソース監視（メモリ、CPU使用量チェック）

3. **launcher**: システム起動とtmuxセッション管理
   - 環境検証（Claude CLI、認証状態の確認）
   - 統合監視画面（6ペイン構成）と個別セッション方式
   - エージェント配置（PO/Manager/Dev1-4）

4. **tmux**: tmux操作の抽象化
   - セッション作成・削除・接続
   - ペイン管理とレイアウト最適化
   - コマンド送信とメッセージ配信

5. **auth**: Claude CLI認証管理
   - 認証状態の確認とバックアップ
   - 設定ファイル（~/.claude/settings.json）の検証

6. **process**: プロセス管理
   - Claude CLIプロセスの監視と終了
   - リソース使用量の追跡

### セッション構成

- **統合監視画面**: 1つのtmuxセッション内に6ペイン（PO, Manager, Dev1-4）
- **個別セッション**: 各エージェント用に独立したtmuxセッション

### 設定ディレクトリ構造
```
~/.claude/
├── settings.json                    # Claude CLI設定
└── claude-code-agents/
    ├── instructions/                # エージェント指示ファイル
    │   ├── po.md
    │   ├── manager.md
    │   └── developer.md
    └── logs/                       # ログファイル
        └── manager.log
```

## 開発のコツ

### テスト実行
- テストは`test/`ディレクトリに分離され、各パッケージ別に整理されています
- モックを使用した単体テストが充実しています
- 統合テストは`launcher.RunIntegrationTests()`で実行されます

### ロギング
- zerologライブラリを使用した構造化ログ
- デバッグモード（--debug）で詳細なログ出力

## コード修正ガイドライン

### 必須品質チェック
**重要**: コード修正後は必ず以下を実行してください：
```bash
make lint
make test
```
- 作業完了前にすべてのlintエラーを修正する
- lintエラーがあるコードは決してコミットしない
- テストは常に実行し、失敗しないことを確認する

### ドキュメント保守
**重要**: 変更時にこのCLAUDE.mdファイルを最新の状態に保ってください：
- モジュールの追加/削除時はアーキテクチャ説明を更新
- Makefileターゲット変更時はコマンドセクションを更新
- 利用可能エージェント変更時はエージェント定義を更新
- セッション処理変更時はセッション検出ロジックを更新
- 新しいGoモジュール追加時は依存関係を更新

### README.mdの保守
**重要**: ユーザに影響のある修正をおこなった場合、README.mdも更新してください。
- 新しい機能追加時はREADME.mdに説明を追加
- README.mdを修正した場合は、該当するdocs/README.en.mdにも英語で追加してください。

これにより、将来のClaude Codeインスタンスがコードベースで正確なガイダンスを得られるようになります。
