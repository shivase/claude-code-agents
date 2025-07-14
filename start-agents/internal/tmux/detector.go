package tmux

import (
	"fmt"
	"os"
)

// TmuxDetectionError はtmux環境内での実行を検出した際のエラー
type TmuxDetectionError struct {
	IsInsideTmux bool
	SessionInfo  string
	PaneInfo     string
}

// Error はエラーメッセージを返す
func (e *TmuxDetectionError) Error() string {
	return "tmux環境内での実行が検出されました"
}

// IsInsideTmux はtmux環境内での実行かどうかを検出する
func IsInsideTmux() (bool, *TmuxDetectionError) {
	tmuxEnv := os.Getenv("TMUX")
	tmuxPane := os.Getenv("TMUX_PANE")

	// TMUX環境変数の存在確認
	if tmuxEnv != "" {
		return true, &TmuxDetectionError{
			IsInsideTmux: true,
			SessionInfo:  tmuxEnv,
			PaneInfo:     tmuxPane,
		}
	}

	// TMUX_PANE環境変数による補足的な確認
	if tmuxPane != "" {
		return true, &TmuxDetectionError{
			IsInsideTmux: true,
			SessionInfo:  tmuxEnv,
			PaneInfo:     tmuxPane,
		}
	}

	return false, nil
}

// PrintErrorMessage はユーザーフレンドリーなエラーメッセージを表示する
func PrintErrorMessage(debugMode bool, err *TmuxDetectionError) {
	if !debugMode {
		// 基本エラーメッセージ
		_, _ = fmt.Fprintf(os.Stderr, "❌ エラー: このコマンドはtmux内から実行できません。\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "tmuxを終了するか、別のターミナルから実行してください。\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "💡 解決方法:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  1. Ctrl+B, D でtmuxをデタッチ\n")
		_, _ = fmt.Fprintf(os.Stderr, "  2. 'exit' でtmuxを終了\n")
		_, _ = fmt.Fprintf(os.Stderr, "  3. 新しいターミナルウィンドウを開く\n")
	} else {
		// デバッグモード時の詳細メッセージ
		_, _ = fmt.Fprintf(os.Stderr, "❌ エラー: tmux環境内での実行が検出されました\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "[デバッグ情報]\n")
		_, _ = fmt.Fprintf(os.Stderr, "  TMUX: %s\n", err.SessionInfo)
		_, _ = fmt.Fprintf(os.Stderr, "  TMUX_PANE: %s\n", err.PaneInfo)

		// セッション名を取得（可能な場合）
		sessionName := os.Getenv("TMUX_SESSION")
		if sessionName != "" {
			_, _ = fmt.Fprintf(os.Stderr, "  セッション名: %s\n", sessionName)
		}

		_, _ = fmt.Fprintf(os.Stderr, "\nこのコマンドはtmuxセッションを管理するため、\n")
		_, _ = fmt.Fprintf(os.Stderr, "tmux環境外から実行する必要があります。\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "💡 解決方法:\n")
		_, _ = fmt.Fprintf(os.Stderr, "  1. Ctrl+B, D でtmuxをデタッチ\n")
		_, _ = fmt.Fprintf(os.Stderr, "  2. 'exit' でtmuxを終了\n")
		_, _ = fmt.Fprintf(os.Stderr, "  3. 新しいターミナルウィンドウを開く\n")
	}
}
