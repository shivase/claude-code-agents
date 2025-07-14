package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/shivase/claude-code-agents/internal/cmd"
	"github.com/stretchr/testify/assert"
)

// テスト用のヘルパー関数：標準出力をキャプチャする
func captureStdout(f func()) (string, error) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan bool)

	go func() {
		defer close(done)
		io.Copy(&buf, r)
	}()

	f()

	w.Close()
	os.Stdout = originalStdout
	<-done

	return buf.String(), nil
}

func TestShowUsage(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// 基本的なヘルプ内容の確認
	t.Run("基本情報の確認", func(t *testing.T) {
		assert.Contains(t, output, "AI並列開発チーム - 統合起動システム")
		assert.Contains(t, output, "使用方法:")
		assert.Contains(t, output, "claude-code-agents")
	})

	// 引数説明の確認
	t.Run("引数説明の確認", func(t *testing.T) {
		assert.Contains(t, output, "引数:")
		assert.Contains(t, output, "セッション名")
		assert.Contains(t, output, "tmuxセッション名")
	})

	// オプション説明の確認
	t.Run("オプション説明の確認", func(t *testing.T) {
		assert.Contains(t, output, "オプション:")
		assert.Contains(t, output, "--reset")
		assert.Contains(t, output, "--verbose")
		assert.Contains(t, output, "--debug")
		assert.Contains(t, output, "--silent")
		assert.Contains(t, output, "--help")

		// ショートオプションの確認
		assert.Contains(t, output, "-v")
		assert.Contains(t, output, "-d")
		assert.Contains(t, output, "-s")
	})

	// 管理コマンド説明の確認
	t.Run("管理コマンド説明の確認", func(t *testing.T) {
		assert.Contains(t, output, "管理コマンド:")
		assert.Contains(t, output, "--list")
		assert.Contains(t, output, "--delete")
		assert.Contains(t, output, "--delete-all")
		assert.Contains(t, output, "--show-config")
		assert.Contains(t, output, "--config")
		assert.Contains(t, output, "--generate-config")
		assert.Contains(t, output, "--init")
		assert.Contains(t, output, "--doctor")
		assert.Contains(t, output, "--force")
	})

	// 使用例の確認
	t.Run("使用例の確認", func(t *testing.T) {
		assert.Contains(t, output, "例:")
		assert.Contains(t, output, "claude-code-agents myproject")
		assert.Contains(t, output, "claude-code-agents ai-team")
		assert.Contains(t, output, "claude-code-agents myproject --reset")
		assert.Contains(t, output, "claude-code-agents myproject --verbose")
		assert.Contains(t, output, "claude-code-agents myproject --silent")
		assert.Contains(t, output, "claude-code-agents --list")
		assert.Contains(t, output, "claude-code-agents --delete myproject")
		assert.Contains(t, output, "claude-code-agents --delete-all")
		assert.Contains(t, output, "claude-code-agents --show-config")
		assert.Contains(t, output, "claude-code-agents --config ai-team")
		assert.Contains(t, output, "claude-code-agents --generate-config")
		assert.Contains(t, output, "claude-code-agents --generate-config --force")
		assert.Contains(t, output, "claude-code-agents --init")
		assert.Contains(t, output, "claude-code-agents --init --force")
		assert.Contains(t, output, "claude-code-agents --doctor")
	})

	// 環境変数説明の確認
	t.Run("環境変数説明の確認", func(t *testing.T) {
		assert.Contains(t, output, "環境変数:")
		assert.Contains(t, output, "VERBOSE=true")
		assert.Contains(t, output, "SILENT=true")
	})
}

func TestShowUsage_OutputFormat(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("出力フォーマットの確認", func(t *testing.T) {
		// 絵文字の確認
		assert.Contains(t, output, "🚀")

		// セクションの区切りの確認
		lines := strings.Split(output, "\n")
		assert.True(t, len(lines) > 10, "十分な行数があること")

		// 空行による適切な区切りがあることを確認
		hasEmptyLines := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				hasEmptyLines = true
				break
			}
		}
		assert.True(t, hasEmptyLines, "セクション間に空行があること")
	})

	t.Run("コマンドライン形式の確認", func(t *testing.T) {
		// 実際のコマンドライン例が正しい形式であることを確認
		commandExamples := []string{
			"claude-code-agents myproject",
			"claude-code-agents ai-team",
			"claude-code-agents myproject --reset",
			"claude-code-agents myproject --verbose",
			"claude-code-agents myproject --silent",
			"claude-code-agents --list",
			"claude-code-agents --delete myproject",
			"claude-code-agents --delete-all",
			"claude-code-agents --show-config",
			"claude-code-agents --config ai-team",
			"claude-code-agents --generate-config",
			"claude-code-agents --generate-config --force",
			"claude-code-agents --init",
			"claude-code-agents --init --force",
			"claude-code-agents --doctor",
		}

		for _, example := range commandExamples {
			assert.Contains(t, output, example, "コマンド例 '%s' が含まれていること", example)
		}
	})
}

func TestShowUsage_JapaneseContent(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("日本語説明の確認", func(t *testing.T) {
		// 主要な日本語説明文の確認
		japaneseTexts := []string{
			"統合起動システム",
			"使用方法",
			"引数",
			"オプション",
			"管理コマンド",
			"例",
			"環境変数",
			"セッション名",
			"詳細ログ出力",
			"デバッグログ出力",
			"サイレントモード",
			"このヘルプを表示",
			"起動中のAIチームセッション一覧を表示",
			"指定したセッションを削除",
			"全てのAIチームセッションを削除",
			"設定値の簡易表示",
			"設定値の詳細表示",
			"設定ファイルのテンプレートを生成",
			"既存ファイルを上書きして生成",
			"システム初期化",
			"既存ファイルを上書きして初期化",
			"システムの健全性チェックを実行",
		}

		for _, text := range japaneseTexts {
			assert.Contains(t, output, text, "日本語テキスト '%s' が含まれていること", text)
		}
	})
}

func TestShowUsage_OptionConsistency(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("オプションの一貫性確認", func(t *testing.T) {
		// ロングオプションとショートオプションの対応確認
		optionPairs := map[string]string{
			"--verbose": "-v",
			"--debug":   "-d",
			"--silent":  "-s",
		}

		for longOpt, shortOpt := range optionPairs {
			assert.Contains(t, output, longOpt, "ロングオプション '%s' が含まれていること", longOpt)
			assert.Contains(t, output, shortOpt, "ショートオプション '%s' が含まれていること", shortOpt)

			// 同じ行にロングとショートが含まれていることを確認
			lines := strings.Split(output, "\n")
			found := false
			for _, line := range lines {
				if strings.Contains(line, longOpt) && strings.Contains(line, shortOpt) {
					found = true
					break
				}
			}
			assert.True(t, found, "オプション '%s' と '%s' が同じ行に含まれていること", longOpt, shortOpt)
		}
	})

	t.Run("必須引数とオプション引数の区別", func(t *testing.T) {
		// 必須引数の表記
		assert.Contains(t, output, "<セッション名>")
		assert.Contains(t, output, "（必須）")

		// オプション引数の表記
		assert.Contains(t, output, "[オプション]")
		assert.Contains(t, output, "[管理コマンド]")
		assert.Contains(t, output, "[名前]")
		assert.Contains(t, output, "[session]")
	})
}

func TestShowUsage_CompleteCoverage(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("全機能の網羅性確認", func(t *testing.T) {
		// parser.goで定義されているすべてのオプションが含まれていることを確認
		allOptions := []string{
			"--help",
			"--verbose",
			"--debug",
			"--silent",
			"--list",
			"--delete",
			"--delete-all",
			"--show-config",
			"--config",
			"--generate-config",
			"--init",
			"--doctor",
			"--reset",
			"--force",
			"-h",
			"-v",
			"-d",
			"-s",
		}

		for _, option := range allOptions {
			assert.Contains(t, output, option, "オプション '%s' がヘルプに含まれていること", option)
		}
	})

	t.Run("環境変数の網羅性確認", func(t *testing.T) {
		envVars := []string{
			"VERBOSE=true",
			"SILENT=true",
		}

		for _, envVar := range envVars {
			assert.Contains(t, output, envVar, "環境変数 '%s' がヘルプに含まれていること", envVar)
		}
	})
}

func TestShowUsage_Structure(t *testing.T) {
	output, err := captureStdout(func() {
		cmd.ShowUsage()
	})

	assert.NoError(t, err)

	t.Run("ヘルプ構造の確認", func(t *testing.T) {
		lines := strings.Split(output, "\n")

		// セクションの順序を確認
		sectionOrder := []string{
			"AI並列開発チーム - 統合起動システム",
			"使用方法:",
			"引数:",
			"オプション:",
			"管理コマンド:",
			"例:",
			"環境変数:",
		}

		lastIndex := -1
		for _, section := range sectionOrder {
			currentIndex := -1
			for i, line := range lines {
				if strings.Contains(line, section) {
					currentIndex = i
					break
				}
			}

			assert.NotEqual(t, -1, currentIndex, "セクション '%s' が見つかること", section)
			assert.Greater(t, currentIndex, lastIndex, "セクション '%s' が正しい順序にあること", section)
			lastIndex = currentIndex
		}
	})

	t.Run("インデントの確認", func(t *testing.T) {
		lines := strings.Split(output, "\n")

		// オプションの説明行が適切にインデントされていることを確認
		optionLines := []string{}
		for _, line := range lines {
			if strings.Contains(line, "--") && !strings.HasPrefix(strings.TrimSpace(line), "claude-code-agents") {
				optionLines = append(optionLines, line)
			}
		}

		assert.Greater(t, len(optionLines), 0, "オプション説明行が存在すること")

		for _, line := range optionLines {
			// 少なくとも2つのスペースでインデントされていることを確認
			assert.True(t, strings.HasPrefix(line, "  "), "オプション行 '%s' が適切にインデントされていること", strings.TrimSpace(line))
		}
	})
}

// パフォーマンステスト
func BenchmarkShowUsage(b *testing.B) {
	// 標準出力を無効化してベンチマークを実行
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = originalStdout
	}()

	for i := 0; i < b.N; i++ {
		cmd.ShowUsage()
	}
}
