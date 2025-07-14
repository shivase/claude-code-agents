package tmux

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// SendInstructionToPaneWithConfig 設定ファイルを使用してインストラクションファイルを送信
func (tm *TmuxManagerImpl) SendInstructionToPaneWithConfig(sessionName, pane, agent, instructionsDir string, config interface{}) error {
	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("📤 設定ベースインストラクション送信開始")

	// インストラクションファイルのパスを決定（設定ファイルから）
	var instructionFile string

	// 設定ファイルからrole別ファイル名を取得
	type InstructionConfig interface {
		GetPOInstructionFile() string
		GetManagerInstructionFile() string
		GetDevInstructionFile() string
	}

	if ic, ok := config.(InstructionConfig); ok {
		switch agent {
		case "po":
			if ic.GetPOInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetPOInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "po.md")
			}
		case "manager":
			if ic.GetManagerInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetManagerInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "manager.md")
			}
		case "dev1", "dev2", "dev3", "dev4":
			if ic.GetDevInstructionFile() != "" {
				instructionFile = filepath.Join(instructionsDir, ic.GetDevInstructionFile())
			} else {
				instructionFile = filepath.Join(instructionsDir, "developer.md")
			}
		default:
			log.Error().Str("agent", agent).Msg("❌ 未知のエージェントタイプ")
			return fmt.Errorf("unknown agent type: %s", agent)
		}
	} else {
		// 設定が提供されていない場合はデフォルトファイル名を使用
		switch agent {
		case "po":
			instructionFile = filepath.Join(instructionsDir, "po.md")
		case "manager":
			instructionFile = filepath.Join(instructionsDir, "manager.md")
		case "dev1", "dev2", "dev3", "dev4":
			instructionFile = filepath.Join(instructionsDir, "developer.md")
		default:
			log.Error().Str("agent", agent).Msg("❌ 未知のエージェントタイプ")
			return fmt.Errorf("unknown agent type: %s", agent)
		}
	}

	log.Info().Str("instruction_file", instructionFile).Msg("📁 設定ベースインストラクションファイルパス決定")

	// インストラクションファイルの存在確認（強化版）
	fileInfo, err := os.Stat(instructionFile)
	if os.IsNotExist(err) {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ インストラクションファイルが存在しません（スキップ）")
		return nil // ファイルが存在しない場合はスキップ（エラーではない）
	}
	if err != nil {
		log.Error().Str("instruction_file", instructionFile).Err(err).Msg("❌ ファイル情報取得エラー")
		return fmt.Errorf("failed to stat instruction file: %w", err)
	}
	if fileInfo.Size() == 0 {
		log.Warn().Str("instruction_file", instructionFile).Msg("⚠️ インストラクションファイルが空です（スキップ）")
		return nil
	}

	log.Info().Str("instruction_file", instructionFile).Int64("file_size", fileInfo.Size()).Msg("✅ 設定ベースファイル存在確認完了")

	// Claude CLI準備完了待機（強化版）
	if err := tm.waitForClaudeReady(sessionName, pane, 10*time.Second); err != nil {
		log.Warn().Str("session", sessionName).Str("pane", pane).Err(err).Msg("⚠️ Claude CLI準備待機タイムアウト（続行）")
	}

	// catコマンドでインストラクションファイルを送信（リトライ機能付き）
	catCommand := fmt.Sprintf("cat \"%s\"", instructionFile)

	for attempt := 1; attempt <= 3; attempt++ {
		log.Info().Str("command", catCommand).Int("attempt", attempt).Msg("📋 設定ベースcatコマンド送信中")

		if err := tm.SendKeysWithEnter(sessionName, pane, catCommand); err != nil {
			log.Warn().Err(err).Int("attempt", attempt).Msg("⚠️ catコマンド送信失敗")
			if attempt == 3 {
				return fmt.Errorf("failed to send instruction file after 3 attempts: %w", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// catコマンド実行完了を待機
		time.Sleep(2 * time.Second)
		log.Info().Int("attempt", attempt).Msg("✅ 設定ベースcatコマンド送信成功")
		break
	}

	// Claude CLI実行のためのEnter送信（最適化版）
	time.Sleep(1 * time.Second)
	log.Info().Msg("🔄 Claude CLI実行のためのEnter送信")

	for i := 0; i < 3; i++ {
		if err := tm.SendKeysToPane(sessionName, pane, "C-m"); err != nil {
			log.Warn().Err(err).Int("attempt", i+1).Msg("⚠️ Enter送信エラー")
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Info().Str("session", sessionName).Str("pane", pane).Str("agent", agent).Msg("✅ 設定ベースインストラクション送信完了")
	return nil
}
