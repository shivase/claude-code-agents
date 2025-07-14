package internal

import (
	"fmt"
	"time"
)

// ヘルパー関数

func IsValidAgent(agent string) bool {
	return ValidAgentNames[agent]
}

func FindAgentByName(name string) *Agent {
	for _, agent := range AvailableAgents {
		if agent.Name == name {
			return &agent
		}
	}
	return nil
}

// メッセージ送信関連のメソッド

func (ms *MessageSender) Send() error {
	target, err := ms.determineTarget()
	if err != nil {
		return err
	}

	if err := ms.sendEnhancedMessage(target); err != nil {
		return err
	}

	ms.displaySummary(target)
	return nil
}

func (ms *MessageSender) determineTarget() (string, error) {
	if HasSession(ms.SessionName) {
		return ms.determineIntegratedTarget()
	}
	return ms.determineIndividualTarget()
}

func (ms *MessageSender) determineIntegratedTarget() (string, error) {
	paneCount, err := GetPaneCount(ms.SessionName)
	if err != nil {
		return "", fmt.Errorf("セッション '%s' の情報取得に失敗しました: %v", ms.SessionName, err)
	}

	if paneCount != IntegratedSessionPaneCount {
		return "", fmt.Errorf("セッション '%s' は統合監視画面形式ではありません", ms.SessionName)
	}

	fmt.Printf("🎯 統合監視画面（%s）を使用してメッセージを送信します\n", ms.SessionName)

	paneIndex := ms.getAgentPaneIndex()
	panes, err := GetPanes(ms.SessionName)
	if err != nil {
		return "", fmt.Errorf("ペイン情報の取得に失敗しました: %v", err)
	}

	if paneIndex < len(panes) {
		target := fmt.Sprintf("%s.%s", ms.SessionName, panes[paneIndex])
		fmt.Printf("📍 %sペイン（ペイン%s）にメッセージを送信\n", ms.Agent, panes[paneIndex])
		return target, nil
	}

	target := fmt.Sprintf("%s.%d", ms.SessionName, paneIndex)
	fmt.Printf("📍 %sペイン（ペイン%d - フォールバック）にメッセージを送信\n", ms.Agent, paneIndex)
	return target, nil
}

func (ms *MessageSender) determineIndividualTarget() (string, error) {
	fmt.Printf("🔄 個別セッション方式（%s）を使用してメッセージを送信します\n", ms.SessionName)

	fullSession := ms.SessionName + "-" + ms.Agent
	if !HasSession(fullSession) {
		return "", fmt.Errorf("セッション '%s' が見つかりません", fullSession)
	}

	return fullSession, nil
}

func (ms *MessageSender) getAgentPaneIndex() int {
	agentPaneMap := map[string]int{
		AgentPO: 0, AgentManager: 1, AgentDev1: 2,
		AgentDev2: 3, AgentDev3: 4, AgentDev4: 5,
	}
	return agentPaneMap[ms.Agent]
}

func (ms *MessageSender) sendEnhancedMessage(target string) error {
	fmt.Printf("📤 送信中: %s へメッセージを送信...\n", ms.Agent)
	fmt.Printf("🎯 対象: %s\n", target)

	// コンテキストリセットが必要な場合
	if ms.ResetContext {
		if err := ms.resetAgentContext(target); err != nil {
			return fmt.Errorf("コンテキストリセットに失敗しました: %v", err)
		}
	}

	// プロンプトクリア
	fmt.Printf("🧹 プロンプトクリア (Ctrl+C)...\n")
	if err := TmuxSendKeys(target, "C-c"); err != nil {
		return fmt.Errorf("プロンプトクリアに失敗しました: %v", err)
	}
	time.Sleep(time.Duration(ClearDelay) * time.Millisecond)

	// 追加のクリア
	fmt.Printf("🧹 追加クリア (Ctrl+U)...\n")
	if err := TmuxSendKeys(target, "C-u"); err != nil {
		return fmt.Errorf("追加クリアに失敗しました: %v", err)
	}
	time.Sleep(time.Duration(AdditionalClearDelay) * time.Millisecond)

	// メッセージ送信
	fmt.Printf("💬 メッセージ送信: \"%s\"\n", ms.Message)
	if err := TmuxSendKeys(target, ms.Message); err != nil {
		return fmt.Errorf("メッセージ送信に失敗しました: %v", err)
	}
	time.Sleep(time.Duration(MessageDelay) * time.Millisecond)

	// Enter押下
	fmt.Printf("⏎ Enter送信 (C-m)...\n")
	if err := TmuxSendKeys(target, "C-m"); err != nil {
		return fmt.Errorf("Enter送信に失敗しました: %v", err)
	}
	time.Sleep(time.Duration(ExecuteDelay) * time.Millisecond)

	fmt.Printf("✅ 送信完了: %s に自動実行されました\n", ms.Agent)
	return nil
}

func (ms *MessageSender) resetAgentContext(target string) error {
	fmt.Printf("🔄 コンテキストリセット開始...\n")

	resetMessage := "前の役割定義やコンテキストは一旦忘れて、新しい指示を待ってください。"

	// リセットメッセージ送信
	fmt.Printf("💭 リセットメッセージ送信: \"%s\"\n", resetMessage)
	if err := TmuxSendKeys(target, resetMessage); err != nil {
		return fmt.Errorf("リセットメッセージ送信に失敗しました: %v", err)
	}
	time.Sleep(time.Duration(MessageDelay) * time.Millisecond)

	// Enter押下
	if err := TmuxSendKeys(target, "C-m"); err != nil {
		return fmt.Errorf("Enter送信に失敗しました: %v", err)
	}
	time.Sleep(time.Duration(ExecuteDelay*3) * time.Millisecond) // より長く待機

	fmt.Printf("✅ コンテキストリセット完了\n")
	return nil
}

func (ms *MessageSender) displaySummary(target string) {
	fmt.Println()
	fmt.Println("🎯 メッセージ詳細:")
	fmt.Printf("   セッション: %s\n", ms.SessionName)
	fmt.Printf("   宛先: %s (%s)\n", ms.Agent, target)
	fmt.Printf("   内容: \"%s\"\n", ms.Message)
	if ms.ResetContext {
		fmt.Printf("   コンテキストリセット: 実行済み\n")
	}
}
