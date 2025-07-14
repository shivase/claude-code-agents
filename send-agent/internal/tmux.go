package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Tmux関連のユーティリティ関数

func GetTmuxSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tmuxセッション一覧の取得に失敗しました: %v", err)
	}

	var sessions []Session
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			sessions = append(sessions, Session{Name: line})
		}
	}

	return sessions, nil
}

func HasSession(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	return cmd.Run() == nil
}

func GetPaneCount(sessionName string) (int, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ペイン数の取得に失敗しました: %v", err)
	}
	return len(strings.Split(strings.TrimSpace(string(output)), "\n")), nil
}

func GetPanes(sessionName string) ([]string, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "#{pane_index}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ペイン一覧の取得に失敗しました: %v", err)
	}

	var panes []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			panes = append(panes, line)
		}
	}
	return panes, nil
}

func ShowPanes(sessionName string) error {
	cmd := exec.Command("tmux", "list-panes", "-t", sessionName, "-F", "  ペイン#{pane_index}: #{pane_title}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ペイン状態の表示に失敗しました: %v", err)
	}
	fmt.Print(string(output))
	return nil
}

func TmuxSendKeys(target, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys)
	return cmd.Run()
}

func DetectDefaultSession() (string, error) {
	sessions, err := GetTmuxSessions()
	if err != nil || len(sessions) == 0 {
		return "", fmt.Errorf("tmuxセッションが見つかりません")
	}

	// 統合監視画面セッション（6ペイン）を優先
	for _, session := range sessions {
		paneCount, err := GetPaneCount(session.Name)
		if err != nil {
			continue
		}
		if paneCount == IntegratedSessionPaneCount {
			return session.Name, nil
		}
	}

	// 潜在的なAIセッションを検出（数字だけのセッション名や短い名前）
	for _, session := range sessions {
		paneCount, err := GetPaneCount(session.Name)
		if err != nil {
			continue
		}
		// 数字だけのセッション名や短い名前、AI関連キーワードのセッションをチェック
		if paneCount >= 1 && (len(session.Name) <= 3 ||
			strings.Contains(session.Name, "ai") ||
			strings.Contains(session.Name, "claude") ||
			strings.Contains(session.Name, "agent")) {
			return session.Name, nil
		}
	}

	// 個別セッション方式のベース名を探す
	individualSessions := map[string]bool{}
	re := regexp.MustCompile(`-(po|manager|dev[1-4])$`)
	for _, session := range sessions {
		if re.MatchString(session.Name) {
			baseName := re.ReplaceAllString(session.Name, "")
			individualSessions[baseName] = true
		}
	}

	if len(individualSessions) > 0 {
		var baseNames []string
		for baseName := range individualSessions {
			baseNames = append(baseNames, baseName)
		}
		sort.Strings(baseNames)
		return baseNames[0], nil
	}

	return "", fmt.Errorf("AIエージェント関連のセッションが見つかりません")
}
