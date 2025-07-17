package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/shivase/claude-code-agents/internal/utils"
)

// ParseArguments parses command line arguments (moved from main.go)
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
			// Set global debug flag (processed by main)
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
				fmt.Println("❌ Error: --delete requires a session name to delete")
				fmt.Println("Usage: ./claude-code-agents --delete [session-name]")
				fmt.Println("Session list: ./claude-code-agents --list")
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
			language := ""

			// 次の引数を確認
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				if args[i+1] == "ja" || args[i+1] == "en" {
					language = args[i+1]
					i++
				} else {
					fmt.Printf("❌ Error: Invalid language '%s'. Use 'ja' or 'en'\n", args[i+1])
					fmt.Println("Usage: ./claude-code-agents --init [ja|en] [--force]")
					os.Exit(1)
				}
			}

			// --forceフラグのチェック
			if i+1 < len(args) && args[i+1] == "--force" {
				forceOverwrite = true
				i++
			}

			// 言語が指定されていない場合はエラー
			if language == "" {
				fmt.Println("❌ Error: Language parameter required")
				fmt.Println("Usage: ./claude-code-agents --init [ja|en] [--force]")
				os.Exit(1)
			}

			if err := InitializeSystemCommand(forceOverwrite, language); err != nil {
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
				fmt.Printf("❌ Error: Unknown option %s\n", arg)
				ShowUsage()
				os.Exit(1)
			} else {
				if sessionName == "" {
					sessionName = arg
				} else {
					fmt.Println("❌ Error: Please specify only one session name")
					ShowUsage()
					os.Exit(1)
				}
			}
		}
		i++
	}

	return sessionName, resetMode, nil
}
