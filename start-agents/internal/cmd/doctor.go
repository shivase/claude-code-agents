package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shivase/claude-code-agents/internal/auth"
	"github.com/shivase/claude-code-agents/internal/config"
	"github.com/shivase/claude-code-agents/internal/utils"
)

// DoctorCommand システムの健全性チェックコマンド
func DoctorCommand() error {
	fmt.Println("🏥 システム健全性チェック")
	fmt.Println("========================")
	fmt.Println()

	var overallStatus = true
	var issues []string

	// 基本環境チェック
	fmt.Println("🔍 基本環境チェック")
	fmt.Println("-------------------")

	// 設定ファイルの読み込み
	fmt.Print("📋 設定ファイル読み込み... ")
	configPath := config.GetDefaultTeamConfigPath()
	configLoader := config.NewTeamConfigLoader(configPath)
	teamConfig, err := configLoader.LoadTeamConfig()
	if err != nil {
		fmt.Printf("❌ 失敗\n")
		fmt.Printf("   エラー: %v\n", err)
		overallStatus = false
		issues = append(issues, "設定ファイルが読み込めません")
	} else {
		fmt.Printf("✅ 成功\n")
	}

	if teamConfig == nil {
		fmt.Println("\n❌ 設定ファイルが読み込めないため、以降のチェックをスキップします")
		return fmt.Errorf("設定ファイルの読み込みに失敗しました")
	}

	fmt.Println()

	// パス存在チェック
	fmt.Println("📂 重要ファイル・ディレクトリチェック")
	fmt.Println("----------------------------------")

	pathChecks := []struct {
		name        string
		path        string
		required    bool
		description string
	}{
		{"Claude CLI実行ファイル", teamConfig.ClaudeCLIPath, true, "Claude CLIの実行に必要"},
		{"インストラクションディレクトリ", teamConfig.InstructionsDir, true, "エージェントの指示ファイル格納"},
		{"作業ディレクトリ", teamConfig.WorkingDir, true, "システムの実行場所"},
		{"設定ディレクトリ", teamConfig.ConfigDir, true, "設定ファイル格納"},
		{"認証バックアップディレクトリ", teamConfig.AuthBackupDir, false, "認証情報のバックアップ"},
		{"ログディレクトリ", filepath.Dir(teamConfig.LogFile), false, "ログファイル格納"},
	}

	for _, check := range pathChecks {
		fmt.Printf("📍 %s: ", check.name)

		expandedPath := utils.ExpandPathSafe(check.path)
		exists := utils.ValidatePath(check.path)

		switch exists {
		case true:
			fmt.Printf("✅ 存在 (%s)\n", utils.FormatPath(check.path))
		case false:
			icon := "❌"
			if !check.required {
				icon = "⚠️"
			}
			fmt.Printf("%s 不在 (%s)\n", icon, utils.FormatPath(check.path))

			if check.required {
				overallStatus = false
				issues = append(issues, fmt.Sprintf("%s が存在しません: %s", check.name, expandedPath))
			} else {
				issues = append(issues, fmt.Sprintf("オプション: %s が存在しません: %s", check.name, expandedPath))
			}
		}
		fmt.Printf("   説明: %s\n", check.description)
		fmt.Println()
	}

	// インストラクションファイル一覧表示
	fmt.Println("📄 インストラクションファイル一覧")
	fmt.Println("----------------------------------")

	instructionsDir := filepath.Join(os.Getenv("HOME"), ".claude", "claude-code-agents", "instructions")
	files, err := os.ReadDir(instructionsDir)
	if err != nil {
		fmt.Printf("📂 インストラクションディレクトリ: %s\n", instructionsDir)
		fmt.Printf("⚠️  ディレクトリが存在しないか、読み取りできません\n")
		fmt.Printf("💡 作成方法: mkdir -p %s\n", instructionsDir)
	} else {
		fmt.Printf("📂 インストラクションディレクトリ: %s\n", instructionsDir)
		if len(files) == 0 {
			fmt.Printf("📝 ファイル数: 0個\n")
			fmt.Printf("💡 役割ファイル例: po.md, manager.md, developer.md\n")
		} else {
			fmt.Printf("📝 ファイル数: %d個\n", len(files))
			for _, file := range files {
				if !file.IsDir() {
					fmt.Printf("   📄 %s\n", file.Name())
				}
			}
		}
	}

	fmt.Println()

	// Claude CLI実行可能性チェック
	fmt.Println("🤖 Claude CLI実行可能性チェック")
	fmt.Println("-----------------------------")

	fmt.Print("🔧 実行権限チェック... ")
	if utils.IsExecutable(utils.ExpandPathSafe(teamConfig.ClaudeCLIPath)) {
		fmt.Printf("✅ 実行可能\n")
	} else {
		fmt.Printf("❌ 実行不可\n")
		overallStatus = false
		issues = append(issues, "Claude CLIに実行権限がありません")
	}

	// Claude認証チェック（OAuth競合防止のため認証テストはスキップ）
	fmt.Print("🔐 Claude認証チェック... ")
	claudeAuth := auth.NewClaudeAuthManager()
	if err := claudeAuth.CheckSettingsFile(); err != nil {
		fmt.Printf("❌ 設定ファイル確認失敗\n")
		fmt.Printf("   エラー: %v\n", err)
		overallStatus = false
		issues = append(issues, "Claude設定ファイルに問題があります")
	} else {
		fmt.Printf("✅ 設定ファイルOK（API認証テストはスキップ）\n")
	}

	fmt.Println()

	// 総合判定
	fmt.Println("📊 診断結果")
	fmt.Println("===========")

	if overallStatus {
		fmt.Println("🎉 システムは正常に動作する準備が整っています！")
		fmt.Println()
		fmt.Println("💡 次のステップ:")
		fmt.Println("   1. claude-code-agents [セッション名] でシステムを起動してください")
		fmt.Println("   2. 各ペインでClaude CLIが正常に動作することを確認してください")
	} else {
		fmt.Println("⚠️ システムに問題が検出されました")
		fmt.Println()
		fmt.Println("🔧 修正が必要な問題:")
		for i, issue := range issues {
			fmt.Printf("   %d. %s\n", i+1, issue)
		}
		fmt.Println()
		fmt.Println("💡 対処方法:")
		fmt.Println("   1. 不足しているファイル・ディレクトリを作成してください")
		fmt.Println("   2. Claude CLIが正しくインストールされているか確認してください")
		fmt.Println("   3. 'claude auth' コマンドで認証を行ってください")
		fmt.Println("   4. 必要に応じて設定ファイルを修正してください")
	}

	fmt.Println()
	fmt.Printf("診断完了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	if !overallStatus {
		return fmt.Errorf("システムに %d 個の問題が検出されました", len(issues))
	}

	return nil
}

// DoctorDetailedCommand 詳細システム診断コマンド（main.goから移動）
func DoctorDetailedCommand() error {
	fmt.Println("🏥 システム診断を開始します...")
	fmt.Println("=====================================")

	var errors []string
	var warnings []string

	// 1. パス検証機能（実行可能ファイル、設定ディレクトリ）
	fmt.Println("\n📁 パス検証機能確認中...")
	if pathErrors := ValidatePathsDetailed(); len(pathErrors) > 0 {
		errors = append(errors, pathErrors...)
	} else {
		fmt.Println("✅ パス検証機能：正常")
	}

	// 2. 設定ファイル確認機能（存在チェック、妥当性検証）
	fmt.Println("\n⚙️ 設定ファイル確認中...")
	if configErrors := ValidateConfigurationDetailed(); len(configErrors) > 0 {
		errors = append(errors, configErrors...)
	} else {
		fmt.Println("✅ 設定ファイル確認機能：正常")
	}

	// 3. Claude認証チェック機能（認証状態、トークン検証）
	fmt.Println("\n🔐 Claude認証状態確認中...")
	if authErrors := ValidateAuthenticationDetailed(); len(authErrors) > 0 {
		warnings = append(warnings, authErrors...)
	} else {
		fmt.Println("✅ Claude認証チェック機能：正常")
	}

	// 4. システム環境確認機能（OS、権限、依存関係）
	fmt.Println("\n🖥️ システム環境確認中...")
	if envErrors := ValidateEnvironmentDetailed(); len(envErrors) > 0 {
		errors = append(errors, envErrors...)
	} else {
		fmt.Println("✅ システム環境確認機能：正常")
	}

	// 5. tmux接続確認（従来機能維持）
	fmt.Println("\n🔧 tmux接続確認中...")
	fmt.Print("📺 tmux可用性... ")
	if _, err := exec.LookPath("tmux"); err != nil {
		errors = append(errors, "tmuxがインストールされていません")
		fmt.Printf("❌ tmuxが見つかりません\n")
	} else {
		fmt.Printf("✅ tmux利用可能\n")
	}

	// 診断結果の詳細表示
	fmt.Println("\n=====================================")
	fmt.Println("🔍 診断結果詳細")
	fmt.Println("=====================================")

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("🎉 システム診断完了 - 全てのチェックが正常です")
		fmt.Printf("診断完了時刻: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	} else {
		if len(errors) > 0 {
			fmt.Println("\n❌ 問題が検出されました：")
			for i, err := range errors {
				fmt.Printf("   %d. %s\n", i+1, err)
			}
			fmt.Println("\n💡 解決策:")
			DisplaySolutionsForErrors(errors)
		}

		if len(warnings) > 0 {
			fmt.Println("\n⚠️ 警告事項：")
			for i, warning := range warnings {
				fmt.Printf("   %d. %s\n", i+1, warning)
			}
			fmt.Println("\n💡 推奨事項:")
			DisplaySolutionsForWarnings(warnings)
		}

		if len(errors) > 0 {
			fmt.Println("\n❌ システムに重要な問題があります。上記の解決策を実行してください。")
			return fmt.Errorf("システム診断で%d個の問題が検出されました", len(errors))
		} else {
			fmt.Println("\n✅ 重要な問題はありませんが、警告事項を確認してください。")
		}
	}

	return nil
}
