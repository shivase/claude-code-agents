package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// expandPath はチルダ（~）をホームディレクトリに展開し、環境変数も展開する
func expandPath(path string) string {
	if path == "" {
		return ""
	}

	// チルダ展開
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path // エラーの場合は元のパスを返す
		}
		path = filepath.Join(homeDir, path[2:])
	} else if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path // エラーの場合は元のパスを返す
		}
		path = homeDir
	}

	// 環境変数の展開
	path = os.ExpandEnv(path)

	// パスをクリーンアップ
	path = filepath.Clean(path)

	return path
}

// validatePath は指定されたパスが存在するかを確認する
func validatePath(path string) bool {
	if path == "" {
		return false
	}

	// パスを展開
	expandedPath := expandPath(path)

	// ファイルまたはディレクトリの存在を確認
	_, err := os.Stat(expandedPath)
	return err == nil
}

// findClaudeExecutable はClaude CLIの実行可能ファイルを指定された優先順位で検索する
func findClaudeExecutable() string {
	// 検索候補のパスリスト（優先順位順）
	candidates := []string{
		"", // which claude の結果用のプレースホルダー
		"~/.claude/local/claude",
		"/usr/local/bin/claude",
		"/opt/homebrew/bin/claude",
		"./claude",
	}

	// 1. which claude の結果を最初に確認
	if whichPath := getWhichClaudeResult(); whichPath != "" {
		candidates[0] = whichPath
	}

	// 各候補をチェック
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		// パスを展開
		expandedPath := expandPath(candidate)

		// ファイルの存在と実行可能性を確認
		if isExecutable(expandedPath) {
			return expandedPath
		}
	}

	// 見つからない場合は空文字を返す
	return ""
}

// getWhichClaudeResult は which claude コマンドの結果を取得する
func getWhichClaudeResult() string {
	cmd := exec.Command("which", "claude")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return ""
	}

	return result
}

// isExecutable はファイルが存在し、実行可能かを確認する
func isExecutable(path string) bool {
	if path == "" {
		return false
	}

	// ファイルの存在を確認
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// ディレクトリの場合は実行可能ファイルではない
	if info.IsDir() {
		return false
	}

	// 実行権限があるかを確認
	mode := info.Mode()
	return mode&0111 != 0 // 実行権限ビットをチェック
}

// getHomeDir はホームディレクトリのパスを取得する（ヘルパー関数）
func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return homeDir
}

// isAbsolutePath は絶対パスかどうかを判定する（ヘルパー関数）
func isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// joinPath は複数のパス要素を結合する（ヘルパー関数）
func joinPath(elements ...string) string {
	return filepath.Join(elements...)
}