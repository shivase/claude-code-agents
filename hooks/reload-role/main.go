package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "hooks",
		Short: "🔄 Claude Code Agents Hook System",
		Long: `🔄 Claude Code Agents Hook System

フックスクリプトによる役割定義の再読み込み機能を提供します。
/reload-role コマンドで指定された役割のmdファイルを再読み込みします。`,
		Example: `  hooks "/reload-role po"
  hooks "/reload-role manager"
  hooks "/reload-role developer"`,
		Args: cobra.ExactArgs(1),
		RunE: executeHook,
	}
)

func main() {
	// TMUX環境で実行されているかチェック
	if isRunningInTmux() {
		fmt.Println("❌ エラー: このコマンドはtmux内では実行できません。")
		fmt.Println("💡 tmuxセッション外で実行してください。")
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}
}

func executeHook(cmd *cobra.Command, args []string) error {
	inputMessage := args[0]

	// /reload-role コマンドかどうかをチェック
	reloadRoleRegex := regexp.MustCompile(`^/reload-role\s+([a-zA-Z]+)`)
	matches := reloadRoleRegex.FindStringSubmatch(inputMessage)

	if len(matches) == 0 {
		// /reload-role コマンドでない場合は何もしない
		return nil
	}

	role := matches[1]

	// 役割が有効かどうかをチェック
	if !isValidRole(role) {
		homeDir, _ := os.UserHomeDir()
		instructionsDir := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions")
		fmt.Println("❌ エラー: 無効な役割です。")
		fmt.Printf("📁 instructionsファイルが見つかりません: %s.%s\n", instructionsDir, role)
		fmt.Println("📝 使用例: /reload-role [role名称]")
		return fmt.Errorf("指定されたrole名に該当するinstructionファイルが存在しません: %s", role)
	}

	// mdファイルのパスを構築
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// ファイルが存在するかチェック
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		fmt.Printf("❌ エラー: %s が見つかりません。\n", mdFile)
		return fmt.Errorf("ファイルが見つかりません: %s", mdFile)
	}

	// ファイルの内容を読み込み
	// #nosec G304 -- mdFileはホームディレクトリ配下の固定パスから構築される
	content, err := os.ReadFile(mdFile)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	// 結果を出力
	fmt.Printf("🔄 %sの役割定義を再読み込み中...\n", role)
	fmt.Println("")
	fmt.Printf("📋 ファイル: %s\n", mdFile)
	fmt.Println("")
	fmt.Println("🔄 前の役割定義をリセットしています...")
	fmt.Println("📖 新しい役割定義を適用します：")
	fmt.Println("----------------------------------------")
	fmt.Print(string(content))
	fmt.Println("----------------------------------------")
	fmt.Println("")
	fmt.Printf("✅ %sの役割定義を正常に再読み込みしました。\n", role)
	fmt.Println("💡 前の役割定義は完全にリセットされ、新しい役割定義のみが適用されます。")

	return nil
}

func isValidRole(role string) bool {
	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// mdファイルのパスを構築
	mdFile := filepath.Join(homeDir, ".claude", "claude-code-agents", "instructions", role+".md")

	// ファイルが存在するかチェック
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// isRunningInTmux checks if the process is running inside a tmux session
func isRunningInTmux() bool {
	// TMUX環境変数が設定されているかチェック
	if os.Getenv("TMUX") != "" {
		return true
	}

	// TERM環境変数にscreenまたはtmuxが含まれているかチェック
	term := os.Getenv("TERM")
	if term == "screen" || term == "screen-256color" || term == "tmux" || term == "tmux-256color" {
		return true
	}

	return false
}
