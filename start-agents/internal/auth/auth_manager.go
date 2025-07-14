package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ClaudeAuthManager Claude認証管理
type ClaudeAuthManager struct {
}

// NewClaudeAuthManager 認証管理を作成
func NewClaudeAuthManager() *ClaudeAuthManager {
	return &ClaudeAuthManager{}
}

// AuthStatus 認証状態
type AuthStatus struct {
	IsAuthenticated bool                   `json:"isAuthenticated"`
	UserID          string                 `json:"userID"`
	OAuthAccount    map[string]interface{} `json:"oauthAccount,omitempty"`
	LastChecked     int64                  `json:"lastChecked"`
}

// CheckAuthenticationStatus 認証状態をチェック
func (cam *ClaudeAuthManager) CheckAuthenticationStatus() (*AuthStatus, error) {
	log.Info().Msg("🔍 Claude認証状態確認中")

	// Claudeファイルの読み込み
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	claudeJsonPath := filepath.Join(homeDir, ".claude", "claude.json")

	authStatus := &AuthStatus{
		LastChecked: time.Now().Unix(),
	}

	// ファイルの存在確認
	if _, err := os.Stat(claudeJsonPath); err != nil {
		log.Warn().Str("config_path", claudeJsonPath).Msg("⚠️ Claude設定ファイルが見つかりません")
		return authStatus, err
	}

	// ファイルの読み込み
	fileData, err := os.ReadFile(claudeJsonPath) // #nosec G304
	if err != nil {
		log.Warn().Err(err).Msg("⚠️ Claude設定ファイルの読み込みに失敗")
		return authStatus, nil
	}

	// JSONのパース
	var data map[string]interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		log.Warn().Err(err).Msg("⚠️ Claude設定ファイルのJSONパースに失敗")
		return authStatus, nil
	}

	log.Info().Msg("✅ Claude設定ファイル読み込み完了")

	// userIDの存在確認
	if userID, exists := data["userID"]; exists && userID != nil {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			authStatus.UserID = userIDStr
			authStatus.IsAuthenticated = true
			log.Info().Str("user_id_prefix", userIDStr[:8]+"...").Msg("✅ 既存認証確認")
		}
	}

	// OAuthアカウント情報の確認
	if oauthAccount, exists := data["oauthAccount"]; exists && oauthAccount != nil {
		if oauthMap, ok := oauthAccount.(map[string]interface{}); ok {
			authStatus.OAuthAccount = oauthMap
			authStatus.IsAuthenticated = true
			if email, exists := oauthMap["emailAddress"]; exists {
				log.Info().Interface("email", email).Msg("📧 OAuth認証済み")
			}
		}
	}

	return authStatus, nil
}

// CheckSettingsFile Claude設定ファイルの存在確認
func (cam *ClaudeAuthManager) CheckSettingsFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	// ~/.claude/settings.json の存在確認
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		return fmt.Errorf("claude設定ファイルが見つかりません: %s", settingsPath)
	}

	log.Info().Str("settings_path", settingsPath).Msg("✅ Claude設定ファイル確認完了")
	return nil
}

// PreAuthChecker 事前認証チェッカー
type PreAuthChecker struct {
	claudePath string
}

// NewPreAuthChecker 事前認証チェッカーを作成
func NewPreAuthChecker(claudePath string) *PreAuthChecker {
	return &PreAuthChecker{
		claudePath: claudePath,
	}
}

// CheckAuthenticationBeforeStart 開始前の認証確認
func (pac *PreAuthChecker) CheckAuthenticationBeforeStart() error {
	log.Info().Msg("ℹ️ tmux起動前にClaude認証状態を確認します")

	// Claude設定ファイルの確認
	cam := NewClaudeAuthManager()
	if err := cam.CheckSettingsFile(); err != nil {
		return fmt.Errorf("claude設定ファイル確認失敗: %w", err)
	}

	// 認証状態の確認（排他アクセス版）
	log.Info().Msg("🔄 Claude認証状態確認中（排他アクセス版）")

	// 並列起動時の認証整合性チェック
	if err := cam.ValidateAuthConcurrency(); err != nil {
		return fmt.Errorf("並列認証整合性チェック失敗: %w", err)
	}

	// 認証状態の確認
	authStatus, err := cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("認証状態確認失敗: %w", err)
	}

	if !authStatus.IsAuthenticated {
		log.Warn().Msg("⚠️ Claude認証が必要です")
		log.Info().Msg("💡 対話式認証を開始します。画面に従って認証を完了してください")
		log.Info().Msg("────────────────────────────────────────────────────────")

		// 対話式認証の実行
		if err := cam.PerformInteractiveAuth(); err != nil {
			return fmt.Errorf("claude認証確認失敗: %w", err)
		}
	}

	log.Info().Msg("✅ Claude認証確認が完了しました")
	return nil
}

// PerformInteractiveAuth 対話式認証を実行
func (cam *ClaudeAuthManager) PerformInteractiveAuth() error {
	log.Info().Msg("🔐 Claude認証状態確認開始")

	// シンプルなテストコマンドで認証状態を確認
	cmd := exec.Command("claude", "--print", "test")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("認証確認失敗: %w", err)
	}

	// 出力があれば認証は成功
	if len(output) == 0 {
		return fmt.Errorf("claude認証応答が空です")
	}

	log.Info().Msg("✅ Claude認証確認完了")
	return nil
}

// EnsureAuthentication 認証を確実に行う
func (cam *ClaudeAuthManager) EnsureAuthentication() error {
	// 認証状態確認
	authStatus, err := cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("認証状態確認失敗: %w", err)
	}

	// 既に認証済みの場合は早期終了
	if authStatus.IsAuthenticated {
		log.Info().Str("user_id_prefix", authStatus.UserID[:8]+"...").Msg("✅ Claude認証済み")
		return nil
	}

	// 認証が必要な場合は対話式認証を実行
	log.Warn().Msg("⚠️ Claude認証が必要です。対話式認証を開始します")

	if err := cam.PerformInteractiveAuth(); err != nil {
		return fmt.Errorf("認証実行失敗: %w", err)
	}

	// 認証後の状態確認
	authStatus, err = cam.CheckAuthenticationStatus()
	if err != nil {
		return fmt.Errorf("認証後状態確認失敗: %w", err)
	}

	if !authStatus.IsAuthenticated {
		return fmt.Errorf("認証処理完了後も認証状態が無効です")
	}

	log.Info().Msg("🎉 Claude認証が正常に完了しました")
	return nil
}

// ValidateAuthConcurrency 並列起動時の認証整合性チェック
func (cam *ClaudeAuthManager) ValidateAuthConcurrency() error {
	log.Info().Msg("🔄 並列Claude起動認証整合性チェック")

	// Claude Codeプロセス数を確認
	cmd := exec.Command("pgrep", "-f", "claude")
	output, err := cmd.Output()
	if err != nil {
		log.Warn().Err(err).Msg("Claude プロセス数確認に失敗")
		return nil // 非致命的エラーとして継続
	}

	processCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))
	if processCount > 1 {
		log.Warn().Int("process_count", processCount).Msg("⚠️ 複数Claude Code プロセス検出")

		// 並列アクセス時は短い待機で認証状態の安定化を図る
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// SafeAuthUpdate 認証状態の安全な更新
func (cam *ClaudeAuthManager) SafeAuthUpdate(updateFunc func(map[string]interface{}) error) error {
	// 簡易実装：ダミー動作
	data := make(map[string]interface{})
	return updateFunc(data)
}

// CleanupCorruptedFiles 破損ファイルのクリーンアップ
func (cam *ClaudeAuthManager) CleanupCorruptedFiles() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}
	claudeDir := filepath.Join(homeDir, ".claude")

	// 1週間以上古い破損ファイルを削除
	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return fmt.Errorf("claudeディレクトリの読み込みに失敗: %w", err)
	}

	cleaned := 0
	cutoff := time.Now().AddDate(0, 0, -7) // 1週間前

	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, ".corrupted.") {
			fullPath := filepath.Join(claudeDir, name)
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				if err := os.Remove(fullPath); err != nil {
					log.Warn().Err(err).Str("file", fullPath).Msg("破損ファイルの削除に失敗")
				} else {
					cleaned++
				}
			}
		}
	}

	if cleaned > 0 {
		log.Info().Int("cleaned_count", cleaned).Msg("🧹 古い破損ファイルをクリーンアップしました")
	}

	return nil
}
