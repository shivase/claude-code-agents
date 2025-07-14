package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shivase/claude-code-agents/internal/utils"
)

// ParseArguments 引数解析関数（main.goから移動）
func ParseArguments(args []string) (string, bool, error) {
	sessionName := ""
	resetMode := false

	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "--help", "-h":
			ShowUsage()
			os.Exit(0)
		case "--verbose", "-v":
			utils.SetVerboseLogging(true)
		case "--debug", "-d":
			// グローバルデバッグフラグを設定（main側で処理）
			return "", false, fmt.Errorf("debug flag processed by main")
		case "--silent", "-s":
			utils.SetSilentMode(true)
		case "--list":
			if err := ListAISessions(); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--delete":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				if err := DeleteAISession(args[i+1]); err != nil {
					return "", false, err
				}
				os.Exit(0)
			} else {
				fmt.Println("❌ エラー: --delete には削除するセッション名が必要です")
				fmt.Println("使用方法: ./claude-code-agents --delete [セッション名]")
				fmt.Println("セッション一覧: ./claude-code-agents --list")
				os.Exit(1)
			}
		case "--delete-all":
			if err := DeleteAllAISessions(); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--show-config":
			if err := DisplayConfigCommand(); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--config":
			sessionName := "ai-teams"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				sessionName = args[i+1]
				i++
			}
			if err := DisplaySessionConfigCommand(sessionName); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--generate-config":
			forceOverwrite := false
			if i+1 < len(args) && args[i+1] == "--force" {
				forceOverwrite = true
				i++
			}
			if err := GenerateConfigCommand(forceOverwrite); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--init":
			forceOverwrite := false
			if i+1 < len(args) && args[i+1] == "--force" {
				forceOverwrite = true
				i++
			}
			if err := InitializeSystemCommand(forceOverwrite); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--doctor":
			if err := DoctorCommand(); err != nil {
				return "", false, err
			}
			os.Exit(0)
		case "--reset":
			resetMode = true
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Printf("❌ エラー: 不明なオプション %s\n", arg)
				ShowUsage()
				os.Exit(1)
			} else {
				if sessionName == "" {
					sessionName = arg
				} else {
					fmt.Println("❌ エラー: セッション名は1つだけ指定してください")
					ShowUsage()
					os.Exit(1)
				}
			}
		}
		i++
	}

	return sessionName, resetMode, nil
}
