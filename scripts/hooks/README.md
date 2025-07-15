# Git Hooks 管理システム

## 概要

このディレクトリは、Cloud Code Agentsプロジェクトで使用されるGit Hooksの統合管理システムです。

## 構成

```
scripts/
└── hooks/
    ├── pre-commit          # Pre-commit hook スクリプト
    └── README.md          # このファイル
```

## 使用方法

### 自動セットアップ（推奨）

プロジェクトのルートディレクトリで以下を実行：

```bash
make hooks-setup
```

このコマンドは以下の処理を実行します：
- 既存のhookをバックアップ
- 適切な方式（シンボリックリンク/コピー）でhookをセットアップ
- プラットフォーム固有の処理を自動実行

### 手動セットアップ

```bash
# Linux/macOS (シンボリックリンク)
ln -sf $(pwd)/scripts/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Windows (コピー)
cp scripts/hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## pre-commitフック機能

### 実行内容
1. **コードフォーマット**: `make fmt`を実行
2. **自動ステージング**: フォーマット変更を自動的にステージング
3. **リンター実行**: `make lint`を実行
4. **エラー時中断**: リントエラーがあればcommitを中断

### 対応プラットフォーム
- ✅ Linux (シンボリックリンク)
- ✅ macOS (シンボリックリンク)
- ✅ Windows (コピー方式)

### 技術仕様
- **言語**: Bash
- **依存関係**: make, git, Makefile
- **エラーハンドリング**: `set -e`による即座終了
- **セキュリティ**: プロジェクトルート検証

## 開発者向け情報

### 新規開発者のセットアップ

1. プロジェクトのクローン
```bash
git clone https://github.com/shivase/cloud-code-agents.git
cd cloud-code-agents
```

2. 依存関係のインストール
```bash
make install
```

3. Git Hooksの確認
```bash
ls -la .git/hooks/pre-commit
```

### フック内容の更新

1. `scripts/hooks/pre-commit`を編集
2. 既存のhookが自動的に更新（シンボリックリンク使用時）
3. コピー方式使用時は`make hooks-setup`を再実行

### トラブルシューティング

**権限エラー**
```bash
chmod +x .git/hooks/pre-commit
```

**フック無効化（一時的）**
```bash
git commit --no-verify -m "commit message"
```

**フック再インストール**
```bash
make hooks-setup
```

## 管理システムの利点

1. **バージョン管理**: フックスクリプトがリポジトリで管理
2. **統一環境**: 全開発者が同じフックを使用
3. **自動更新**: シンボリックリンクによる自動同期
4. **プラットフォーム対応**: OS固有の処理を自動選択
5. **保守性**: 中央集約管理による更新の容易さ

## 注意事項

- フックは`.git/hooks/`ディレクトリに配置されるため、リポジトリには含まれません
- 新規開発者は`make hooks-setup`の実行が必要です
- フックの無効化は`--no-verify`オプションで可能です
- バックアップは`.git/hooks/pre-commit.backup`に保存されます