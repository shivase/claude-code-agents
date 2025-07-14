package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shivase/cloud-code-agents/send-agent/internal"
)

// コマンド定義
var (
	rootCmd = &cobra.Command{
		Use:   "send-agent [agent] [message]",
		Short: "🚀 AIエージェント メッセージ送信システム",
		Long: `🚀 AIエージェント メッセージ送信システム

tmuxセッション上のAIエージェントにメッセージを送信するツールです。
統合監視画面および個別セッション方式の両方に対応しています。

利用可能エージェント:
  po      - プロダクトオーナー（製品責任者）
  manager - プロジェクトマネージャー（柔軟なチーム管理）
  dev1    - 実行エージェント1（柔軟な役割対応）
  dev2    - 実行エージェント2（柔軟な役割対応）
  dev3    - 実行エージェント3（柔軟な役割対応）
  dev4    - 実行エージェント4（柔軟な役割対応）`,
		Example: `  send-agent --session myproject manager "新しいプロジェクトを開始してください"
  send-agent --session ai-team dev1 "【マーケティング担当として】市場調査を実施してください"
  send-agent --reset dev1 "【データアナリストとして】レポートを作成してください"
  send-agent manager "メッセージ"  (デフォルトセッション使用)
  send-agent list myproject      (myprojectセッションのエージェント一覧)
  send-agent list-sessions       (全セッション一覧表示)`,
		Args: cobra.ExactArgs(2),
		RunE: executeMainCommand,
	}

	listCmd = &cobra.Command{
		Use:   "list [session-name]",
		Short: "指定したセッションのエージェント一覧を表示",
		Args:  cobra.ExactArgs(1),
		RunE:  executeListCommand,
	}

	listSessionsCmd = &cobra.Command{
		Use:   "list-sessions",
		Short: "利用可能な全セッション一覧を表示",
		RunE:  executeListSessionsCommand,
	}
)

func init() {
	rootCmd.Flags().StringP("session", "s", "", "指定したセッション名を使用")
	rootCmd.Flags().BoolP("reset", "r", false, "前の役割定義をクリアして新しい指示を送信")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(listSessionsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		os.Exit(1)
	}
}

// コマンド実行関数
func executeMainCommand(cmd *cobra.Command, args []string) error {
	agent := args[0]
	message := args[1]
	sessionName, _ := cmd.Flags().GetString("session")
	resetContext, _ := cmd.Flags().GetBool("reset")

	if !internal.IsValidAgent(agent) {
		return fmt.Errorf("無効なエージェント名 '%s'", agent)
	}

	if sessionName == "" {
		detectedSession, err := internal.DetectDefaultSession()
		if err != nil {
			return fmt.Errorf("利用可能なAIエージェントセッションが見つかりません\n💡 セッション一覧: %s list-sessions\n💡 新しいセッション作成: start-ai-agent [セッション名]", cmd.Root().Name())
		}
		sessionName = detectedSession
		fmt.Printf("🔍 デフォルトセッション '%s' を使用します\n", sessionName)
	}

	sender := &internal.MessageSender{
		SessionName:  sessionName,
		Agent:        agent,
		Message:      message,
		ResetContext: resetContext,
	}

	return sender.Send()
}

func executeListCommand(cmd *cobra.Command, args []string) error {
	sessionName := args[0]
	manager := &internal.SessionManager{}
	return manager.ShowAgentsForSession(sessionName)
}

func executeListSessionsCommand(cmd *cobra.Command, args []string) error {
	manager := &internal.SessionManager{}
	return manager.ListAllSessions()
}
