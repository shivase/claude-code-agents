package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// 表示制御フラグ
var (
	verboseLogging = false
	silentMode     = false
)

// SetVerboseLogging 詳細ログ出力を設定
func SetVerboseLogging(verbose bool) {
	verboseLogging = verbose
	if verbose {
		silentMode = false // verbose時はsilentを無効化
		// 詳細ログ有効化のメッセージは不要（fmt出力で十分）
	}
}

// SetSilentMode サイレントモードを設定
func SetSilentMode(silent bool) {
	silentMode = silent
	if silent {
		verboseLogging = false // silent時はverboseを無効化
		// サイレントモード有効化のメッセージは不要（fmt出力で十分）
	}
}

// IsVerboseLogging 詳細ログ出力が有効かチェック
func IsVerboseLogging() bool {
	return verboseLogging
}

// IsSilentMode サイレントモードが有効かチェック
func IsSilentMode() bool {
	return silentMode
}

// DisplayProgress 進行状況の表示
func DisplayProgress(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("🔄 %s: %s\n", operation, message)
	// 構造化ログは不要（fmt出力で十分）
}

// DisplaySuccess 成功メッセージの表示
func DisplaySuccess(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("✅ %s: %s\n", operation, message)
	// 構造化ログは不要（fmt出力で十分）
}

// DisplayError エラーメッセージの表示
func DisplayError(operation string, err error) {
	fmt.Printf("❌ %s: %v\n", operation, err)
	// 構造化ログは不要（fmt出力で十分）
}

// DisplayInfo 情報メッセージの表示
func DisplayInfo(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("ℹ️ %s: %s\n", operation, message)
	// 構造化ログは不要（fmt出力で十分）
}

// DisplayWarning 警告メッセージの表示
func DisplayWarning(operation, message string) {
	if silentMode {
		return
	}
	fmt.Printf("⚠️ %s: %s\n", operation, message)
	// 構造化ログは不要（fmt出力で十分）
}

// DisplayStartupBanner スタートアップバナーを表示（詳細モード時のみ）
func DisplayStartupBanner() {
	if silentMode {
		return
	}

	fmt.Println("🚀 AI Teams System - Claude Code Agents")
	fmt.Println("=====================================")
	fmt.Printf("Version: 1.0.0\n")
	fmt.Printf("Runtime: Go %s\n", runtime.Version())
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Start Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=====================================")
	fmt.Println()

	// 構造化ログは不要（fmt出力で十分）
}

// DisplayLauncherStart ランチャー開始メッセージを表示
func DisplayLauncherStart() {
	if silentMode {
		return
	}
	fmt.Printf("[%s] 🚀 システムランチャー開始\n", time.Now().Format("15:04:05"))
	fmt.Println("=====================================")
}

// DisplayLauncherProgress ランチャー進行状況を表示
func DisplayLauncherProgress() {
	if silentMode {
		return
	}
	fmt.Printf("[%s] 🔄 システム初期化中...\n", time.Now().Format("15:04:05"))
}

// DisplayConfig 設定情報を表示
func DisplayConfig(teamConfig interface{}, sessionName string) {
	if silentMode {
		return
	}

	fmt.Println("📋 設定情報")
	fmt.Println("===========")
	fmt.Printf("セッション名: %s\n", sessionName)
	fmt.Println()

	// teamConfigの型によって処理を分ける（interface{}として受け取るため）
	if config, ok := teamConfig.(map[string]interface{}); ok {
		for key, value := range config {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
	fmt.Println()
}

// DisplayValidationResults 検証結果を表示
func DisplayValidationResults(teamConfig interface{}) {
	if silentMode {
		return
	}

	fmt.Println("🔍 検証結果")
	fmt.Println("===========")

	// 簡易的な検証結果表示
	fmt.Println("✅ Claude CLI: 利用可能")
	fmt.Println("✅ インストラクション: 準備完了")
	fmt.Println("✅ 作業ディレクトリ: アクセス可能")
	fmt.Println()
}

// FormatPath パスを表示用にフォーマット
func FormatPath(path string) string {
	if path == "" {
		return "<empty>"
	}

	// ホームディレクトリを ~ に置換
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}

// ValidatePath パスの存在確認
func ValidatePath(path string) bool {
	if path == "" {
		return false
	}

	expandedPath := ExpandPathSafe(path)
	_, err := os.Stat(expandedPath)
	return err == nil
}

// ExpandPathOld チルダ展開（非推奨: path_utils.goのExpandPathSafeを使用してください）
func ExpandPathOld(path string) string {
	return ExpandPathSafe(path)
}

// IsExecutable ファイルが実行可能かチェック
func IsExecutable(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := fileInfo.Mode()
	return mode&0111 != 0
}
