# Cloud Code Agents

[README in English is here](./docs/README.en.md) 

tmux上でClaude Code AIエージェントを並列実行する統合開発環境システムです。
Product Ownerに指示を与えることで、Managerが適切に作業を解析し、各種Dev Roleのエージェントが並列で処理をおこなうようになります。

## 概要

このプロジェクトは、複数のAIエージェントを並列で実行し、チーム開発を効率化するためのツールセットです。以下の2つの主要コンポーネントで構成されています：

![screen_shot](docs/screen_shot.png)

- **start-agents**: AIエージェントセッションを起動・管理するメインシステム
- **send-agent**: 起動中のエージェントにメッセージを送信するクライアントツール

## 🚀 使用方法

### 事前作業

起動に必要な各種環境情報を`--init`コマンドで作成します。
ファイルはデフォルトでは`~/.claude/claude-code-agents/agents.json`に保存されます。

```bash
git clone https://github.com/shivase/claude-code-agents.git
cd claude-code-agents
# install start-agents and send-agent to /usr/local/bin
make install

# 設定初期化
claude-code-agents --init

# システム診断実行
claude-code-agents --doctor

# 設定ファイルの情報を見たい場合
claude-code-agents --show-config
```

### エージェントの制約を設定する。

このプログラムで起動するclaude codeは`dangerously-skip-permissions`をONにして起動しています。
そのため、エージェントに制約を与えておかなければ、かなり勝手に動いてしまい不都合があります。

[settings.json](./docs/settings.json)を参考に、`~/.claude/claude-code-agents/settings.json`の、`allow`と`deny`の設定をおこなってください。

#### エージェントの起動

claude codeは起動ディレクトリに依存して動作します。

```bash
# 対象のプロジェクトフォルダに移動
cd [プロジェクトフォルダ]
# セッション名を指定して起動してください
claude-code-agents [session_name]
```

**起動されるエージェント：**
- `po`: プロダクトオーナー（全体統括）, 左上pane
- `manager`: プロジェクトマネージャー（チーム管理）, 左下pane
- `dev1-dev4`: 実行エージェント（柔軟な役割対応） , 右側pane

#### 各エージェントの定義ファイル

各種エージェントの動作定義は`~/.claude/claude-code-agents/instructions`に保存されています。
名前は `<role>.md`の形式で保存されますので、 自身の環境に合わせて任意に変更した上で、再度アプリを立ち上げなおして下さい。

## FAQ

### Q: 起動をもっと早くできないか？

元々このプログラムはbashで作成していました。
その時の起動時間を短縮するために、各エージェントの起動を並列化する目的でGoで作り直した背景があります。
しかし、claude codeの認証情報は一つのファイルで管理しており、同時起動するとこのファイルが壊れることが多々あり、各起動したpaneで認証入力が必要になりました。

そのため、現状は一つずつタイマーでずらしながら起動しており遅い状態です。（高速化したい･･･

### Q: PO/Manager/Devが適切に別のRoleにデータを投げなくなった。

これは、会話が多量になった場合や、たまに対処に起きる場合があります。
以下のように、Roleファイルを再読み込みさせると治ることがあります。

```bash
cat "~/.claude/claude-code-agents/instructions/developer.md"
```

### Q: PO/Managerが自身でコード生成するようになった

「あなたはPOです。自身で作業するのではなく指示をしてください」などといった指示を与えると治ります。

### Q: DeveloperがManagerに作業報告していない

これも会話量が増えてくると起きる時があります。

「作業が完了していれば、send-agentでmanagerに適切に報告して下さい」

という文言を入れてあげると治ります。もしくは Roleファイルを再読み込みさせると良いでしょう。

## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🤝 貢献

プロジェクトへの貢献を歓迎します。

- [Issues](https://github.com/shivase/cloud-code-agents/issues) - バグ報告・機能要求
- [Pull Requests](https://github.com/shivase/cloud-code-agents/pulls) - コード貢献
- [Discussions](https://github.com/shivase/cloud-code-agents/discussions) - 質問・議論
