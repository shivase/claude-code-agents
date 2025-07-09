package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	// グローバル設定フラグ
	verbose   bool
	logLevel  string
	configDir string
	
	// 初期化管理
	mainInitialized bool
	mainInitMutex   sync.Mutex
)

// showUsage ヘルプメッセージを表示
func showUsage() {
	fmt.Println("🚀 AI並列開発チーム - 統合起動システム")
	fmt.Println("")
	fmt.Println("使用方法:")
	fmt.Println("  ./claude-code-agents <セッション名> [オプション]")
	fmt.Println("  ./claude-code-agents [管理コマンド]")
	fmt.Println("")
	fmt.Println("引数:")
	fmt.Println("  セッション名      tmuxセッション名（必須）")
	fmt.Println("  ")
	fmt.Println("オプション:")
	fmt.Println("  --reset          既存セッションを削除して再作成")
	fmt.Println("  --individual     個別セッション方式で起動（統合監視画面なし）")
	fmt.Println("  --verbose, -v    詳細ログ出力を有効化")
	fmt.Println("  --silent, -s     サイレントモード（ログ出力を最小化）")
	fmt.Println("  --help           このヘルプを表示")
	fmt.Println("")
	fmt.Println("管理コマンド:")
	fmt.Println("  --list             起動中のAIチームセッション一覧を表示")
	fmt.Println("  --delete [名前]    指定したセッションを削除")
	fmt.Println("  --delete-all       全てのAIチームセッションを削除")
	fmt.Println("  --show-config      設定値の簡易表示")
	fmt.Println("  --config [session] 設定値の詳細表示")
	fmt.Println("  --generate-config  設定ファイルのテンプレートを生成")
	fmt.Println("    --force          既存ファイルを上書きして生成")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  claude-code-agents myproject               # myprojectセッションで統合監視画面起動")
	fmt.Println("  claude-code-agents ai-team                 # ai-teamセッションで統合監視画面起動")
	fmt.Println("  claude-code-agents myproject --reset       # myprojectセッションを再作成")
	fmt.Println("  claude-code-agents myproject --individual  # myprojectで個別セッション方式起動")
	fmt.Println("  claude-code-agents myproject --verbose     # 詳細ログ付きで起動")
	fmt.Println("  claude-code-agents myproject --silent      # サイレントモードで起動")
	fmt.Println("  claude-code-agents --list                    # セッション一覧表示")
	fmt.Println("  claude-code-agents --delete myproject        # myprojectセッションを削除")
	fmt.Println("  claude-code-agents --delete-all              # 全セッション削除")
	fmt.Println("  claude-code-agents --show-config             # 設定値の簡易表示")
	fmt.Println("  claude-code-agents --config ai-team          # ai-teamセッションの設定値詳細表示")
	fmt.Println("  claude-code-agents --generate-config         # 設定ファイルのテンプレートを生成")
	fmt.Println("  claude-code-agents --generate-config --force # 既存ファイルを上書きして生成")
	fmt.Println("")
	fmt.Println("環境変数:")
	fmt.Println("  VERBOSE=true       詳細ログ出力を有効化")
	fmt.Println("  SILENT=true        サイレントモードを有効化")
	fmt.Println("")
}

// listAISessions セッション一覧表示機能
func listAISessions() error {
	fmt.Println("🤖 セッション一覧")
	fmt.Println("==================================")
	
	// tmux list-sessions を実行して、セッション名のみを取得
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 現在起動中のAIチームセッションはありません")
			return nil
		}
		return fmt.Errorf("tmuxセッション取得エラー: %v", err)
	}
	
	sessionsOutput := string(output)
	sessions := strings.Fields(sessionsOutput)
	
	if len(sessions) == 0 {
		fmt.Println("📭 現在起動中のセッションはありません")
	} else {
		fmt.Printf("🚀 起動中のセッション: %d個\n", len(sessions))
		for i, session := range sessions {
			fmt.Printf("  %d. %s\n", i+1, session)
		}
	}
	
	return nil
}

// deleteAISession 指定したセッションを削除
func deleteAISession(sessionName string) error {
	if sessionName == "" {
		fmt.Println("❌ エラー: 削除するセッション名を指定してください")
		fmt.Println("使用方法: ./claude-code-agents --delete [セッション名]")
		fmt.Println("セッション一覧: ./claude-code-agents --list")
		return fmt.Errorf("セッション名が指定されていません")
	}
	
	fmt.Printf("🗑️ セッション削除: %s\n", sessionName)
	
	// セッションの存在確認
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ セッション '%s' は存在しません\n", sessionName)
		return nil
	}
	
	// セッションを削除
	cmd = exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("セッション削除エラー: %v", err)
	}
	
	fmt.Printf("✅ セッション '%s' を削除しました\n", sessionName)
	return nil
}

// deleteAllAISessions 全セッションを削除
func deleteAllAISessions() error {
	fmt.Println("🗑️ 全AIチームセッション削除")
	fmt.Println("==============================")
	
	// 現在のセッション一覧を取得
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "no server running") {
			fmt.Println("📭 現在起動中のAIチームセッションはありません")
			return nil
		}
		return fmt.Errorf("tmuxセッション取得エラー: %v", err)
	}
	
	sessions := strings.Fields(string(output))
	aiSessions := []string{}
	
	// AIチーム関連のセッションを抽出
	for _, session := range sessions {
		if strings.Contains(session, "ai-") || strings.Contains(session, "claude-") || 
		   strings.Contains(session, "dev-") || strings.Contains(session, "agent-") {
			aiSessions = append(aiSessions, session)
		}
	}
	
	if len(aiSessions) == 0 {
		fmt.Println("📭 削除対象のAIチームセッションはありません")
		return nil
	}
	
	fmt.Printf("🎯 削除対象セッション: %d個\n", len(aiSessions))
	for i, session := range aiSessions {
		fmt.Printf("  %d. %s\n", i+1, session)
	}
	
	// 各セッションを削除
	deletedCount := 0
	for _, session := range aiSessions {
		cmd := exec.Command("tmux", "kill-session", "-t", session)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️ セッション '%s' の削除に失敗: %v\n", session, err)
		} else {
			deletedCount++
			fmt.Printf("✅ セッション '%s' を削除しました\n", session)
		}
	}
	
	fmt.Printf("\n🎉 %d個のAIチームセッションを削除しました\n", deletedCount)
	return nil
}

// 引数解析関数
func parseArguments(args []string) (string, bool, bool, error) {
	sessionName := ""
	resetMode := false
	individualMode := false
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		switch arg {
		case "--help", "-h":
			showUsage()
			os.Exit(0)
		case "--verbose", "-v":
			SetVerboseLogging(true)
		case "--silent", "-s":
			SetSilentMode(true)
		case "--list":
			if err := listAISessions(); err != nil {
				return "", false, false, err
			}
			os.Exit(0)
		case "--delete":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				if err := deleteAISession(args[i+1]); err != nil {
					return "", false, false, err
				}
				os.Exit(0)
			} else {
				fmt.Println("❌ エラー: --delete には削除するセッション名が必要です")
				fmt.Println("使用方法: ./claude-code-agents --delete [セッション名]")
				fmt.Println("セッション一覧: ./claude-code-agents --list")
				os.Exit(1)
			}
		case "--delete-all":
			if err := deleteAllAISessions(); err != nil {
				return "", false, false, err
			}
			os.Exit(0)
		case "--show-config":
			if err := showConfigCommand("ai-teams"); err != nil {
				return "", false, false, err
			}
			os.Exit(0)
		case "--config":
			sessionName := "ai-teams"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				sessionName = args[i+1]
				i++
			}
			if err := displayConfigCommand(sessionName, true); err != nil {
				return "", false, false, err
			}
			os.Exit(0)
		case "--generate-config":
			forceOverwrite := false
			if i+1 < len(args) && args[i+1] == "--force" {
				forceOverwrite = true
				i++
			}
			if err := generateConfigCommand(forceOverwrite); err != nil {
				return "", false, false, err
			}
			os.Exit(0)
		case "--reset":
			resetMode = true
		case "--individual":
			individualMode = true
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Printf("❌ エラー: 不明なオプション %s\n", arg)
				showUsage()
				os.Exit(1)
			} else {
				if sessionName == "" {
					sessionName = arg
				} else {
					fmt.Println("❌ エラー: セッション名は1つだけ指定してください")
					showUsage()
					os.Exit(1)
				}
			}
		}
		i++
	}
	
	return sessionName, resetMode, individualMode, nil
}

// initializeMainSystem メインシステムの初期化処理統合化
func initializeMainSystem() {
	mainInitMutex.Lock()
	defer mainInitMutex.Unlock()
	
	// 既に初期化されている場合はスキップ
	if mainInitialized {
		return
	}
	
	// ログシステムの初期化
	initLogger()
	
	// 共通設定の初期化
	_ = GetCommonConfig()
	
	// 初期化フラグを設定
	mainInitialized = true
}

// main 関数
func main() {
	args := os.Args[1:]
	
	// 引数解析
	sessionName, resetMode, individualMode, err := parseArguments(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
	
	// セッション名が指定されていない場合はヘルプを表示
	if sessionName == "" {
		fmt.Println("❌ エラー: セッション名を指定してください")
		fmt.Println("")
		showUsage()
		os.Exit(1)
	}
	
	// セッション名のバリデーション
	if !isValidSessionName(sessionName) {
		fmt.Println("❌ エラー: セッション名は英数字、ハイフン、アンダースコアのみ使用可能です")
		os.Exit(1)
	}
	
	// 初期化処理の統合化
	initializeMainSystem()
	
	// ディレクトリ解決器の初期化
	if err := InitializeDirectoryResolver(); err != nil {
		fmt.Printf("エラー: ディレクトリ解決器の初期化に失敗: %v\n", err)
		// エラーがあっても続行する
	}
	
	// 環境変数からログレベルを設定
	if verboseMode := os.Getenv("VERBOSE"); verboseMode == "true" || verboseMode == "1" {
		SetVerboseLogging(true)
	}
	if silentMode := os.Getenv("SILENT"); silentMode == "true" || silentMode == "1" {
		SetSilentMode(true)
	}
	
	// スタートアップバナー表示（詳細モードのみ）
	if IsVerboseLogging() {
		displayStartupBanner()
	}
	
	// 設定ファイルの読み込み
	configPath := GetTeamConfigPath()
	teamConfig, err := LoadTeamConfig()
	
	// ディレクトリ情報の表示（詳細モードのみ）
	if IsVerboseLogging() {
		resolver := GetGlobalDirectoryResolver()
		resolver.DisplayDirectoryInfo()
	}
	if err != nil {
		displayError("設定ファイル読み込み", err)
		os.Exit(1)
	}
	
	// 簡素化された形式で設定情報を表示
	displayClaudePath(teamConfig.ClaudeCLIPath)
	displayConfigFileLoaded(configPath, "")
	displaySessionName(sessionName)
	
	// 詳細モードのみで設定情報とパス検証結果を表示
	if IsVerboseLogging() {
		displayConfig(teamConfig, sessionName)
		displayValidationResults(teamConfig)
	}
	
	// システム起動
	layout := "integrated"
	if individualMode {
		layout = "individual"
	}
	
	if err := startSystemWithLauncher(sessionName, layout, resetMode); err != nil {
		displayError("システム起動エラー", err)
		os.Exit(1)
	}
}

// isValidSessionName セッション名のバリデーション
func isValidSessionName(name string) bool {
	if name == "" {
		return false
	}
	
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}
	return true
}

