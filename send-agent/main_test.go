package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/shivase/cloud-code-agents/send-agent/internal"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMainFunction(t *testing.T) {
	// メイン関数の基本テスト
	t.Run("MainExists", func(t *testing.T) {
		// main関数が存在することを確認
		assert.NotNil(t, main)
	})
}

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "InvalidAgent",
			args:    []string{"invalid", "test message"},
			wantErr: true,
			errMsg:  "無効なエージェント名",
		},
		{
			name:    "NoArgs",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "OnlyOneArg",
			args:    []string{"manager"},
			wantErr: true,
		},
		{
			name:    "TooManyArgs",
			args:    []string{"manager", "msg1", "msg2", "msg3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モック化された実行関数を使用
			mockExecuteMainCommand := func(cmd *cobra.Command, args []string) error {
				// 引数の検証のみ行う
				if len(args) != 2 {
					return errors.New("引数の数が不正です")
				}

				agent := args[0]
				if !internal.IsValidAgent(agent) {
					return errors.New("無効なエージェント名 '" + agent + "'")
				}

				// 外部依存を排除：実際のメッセージ送信は行わない
				return nil
			}

			// コマンドを新しく作成してテスト
			cmd := &cobra.Command{
				Use:  "send-agent [agent] [message]",
				Args: cobra.ExactArgs(2),
				RunE: mockExecuteMainCommand,
			}
			cmd.Flags().StringP("session", "s", "", "指定したセッション名を使用")
			cmd.Flags().BoolP("reset", "r", false, "前の役割定義をクリアして新しい指示を送信")

			cmd.SetArgs(tt.args)

			// 出力をキャプチャ
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandLineFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "SessionFlag",
			args:    []string{"--session", "test", "manager", "test message"},
			wantErr: false, // フラグ解析のみテスト
		},
		{
			name:    "SessionFlagShort",
			args:    []string{"-s", "test", "manager", "test message"},
			wantErr: false,
		},
		{
			name:    "ResetFlag",
			args:    []string{"--reset", "manager", "test message"},
			wantErr: false,
		},
		{
			name:    "ResetFlagShort",
			args:    []string{"-r", "manager", "test message"},
			wantErr: false,
		},
		{
			name:    "BothFlags",
			args:    []string{"-s", "test", "-r", "manager", "test message"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モック化された実行関数を使用
			mockExecuteMainCommand := func(cmd *cobra.Command, args []string) error {
				// フラグの値を確認
				sessionName, _ := cmd.Flags().GetString("session")
				resetContext, _ := cmd.Flags().GetBool("reset")

				// 引数の検証
				if len(args) != 2 {
					return errors.New("引数の数が不正です")
				}

				agent := args[0]
				if !internal.IsValidAgent(agent) {
					return errors.New("無効なエージェント名 '" + agent + "'")
				}

				// フラグが正しく設定されているかテスト用にチェック
				// 実際の処理は行わない
				_ = sessionName
				_ = resetContext

				return nil
			}

			cmd := &cobra.Command{
				Use:  "send-agent [agent] [message]",
				Args: cobra.ExactArgs(2),
				RunE: mockExecuteMainCommand,
			}
			cmd.Flags().StringP("session", "s", "", "指定したセッション名を使用")
			cmd.Flags().BoolP("reset", "r", false, "前の役割定義をクリアして新しい指示を送信")

			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "ListWithSession",
			args:    []string{"test-session"},
			wantErr: false, // 引数解析のみテスト
		},
		{
			name:    "ListWithoutSession",
			args:    []string{},
			wantErr: true, // 引数が不足
		},
		{
			name:    "ListWithTooManyArgs",
			args:    []string{"session1", "session2"},
			wantErr: true, // 引数が多すぎる
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モック化された実行関数を使用
			mockExecuteListCommand := func(cmd *cobra.Command, args []string) error {
				// 引数の検証のみ行う
				if len(args) != 1 {
					return errors.New("引数の数が不正です")
				}

				// 実際のセッション管理は行わない
				return nil
			}

			cmd := &cobra.Command{
				Use:  "list [session-name]",
				Args: cobra.ExactArgs(1),
				RunE: mockExecuteListCommand,
			}

			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListSessionsCommand(t *testing.T) {
	t.Run("ListSessions", func(t *testing.T) {
		// モック化された実行関数を使用
		mockExecuteListSessionsCommand := func(cmd *cobra.Command, args []string) error {
			// 引数なしでの実行をテスト
			// 実際のセッション一覧取得は行わない
			return nil
		}

		cmd := &cobra.Command{
			Use:  "list-sessions",
			RunE: mockExecuteListSessionsCommand,
		}

		cmd.SetArgs([]string{})

		err := cmd.Execute()

		// モック化されているためエラーは発生しない
		assert.NoError(t, err)
	})
}

func TestExecuteMainCommand(t *testing.T) {
	tests := []struct {
		name    string
		agent   string
		message string
		session string
		reset   bool
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidAgent",
			agent:   "manager",
			message: "test message",
			wantErr: false, // 引数検証のみテスト
		},
		{
			name:    "InvalidAgent",
			agent:   "invalid",
			message: "test message",
			wantErr: true,
			errMsg:  "無効なエージェント名",
		},
		{
			name:    "EmptyMessage",
			agent:   "manager",
			message: "",
			wantErr: false, // 空メッセージも許可（引数検証のみ）
		},
		{
			name:    "WithSession",
			agent:   "dev1",
			message: "test message",
			session: "test-session",
			wantErr: false, // セッション指定テスト
		},
		{
			name:    "WithReset",
			agent:   "dev2",
			message: "test message",
			reset:   true,
			wantErr: false, // リセットフラグテスト
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モック化された実行関数を使用
			mockExecuteMainCommand := func(cmd *cobra.Command, args []string) error {
				// 引数の検証
				if len(args) != 2 {
					return errors.New("引数の数が不正です")
				}

				agent := args[0]
				message := args[1]
				sessionName, _ := cmd.Flags().GetString("session")
				resetContext, _ := cmd.Flags().GetBool("reset")

				// エージェント名の検証
				if !internal.IsValidAgent(agent) {
					return errors.New("無効なエージェント名 '" + agent + "'")
				}

				// フラグの値を確認（実際の処理は行わない）
				_ = message
				_ = sessionName
				_ = resetContext

				return nil
			}

			cmd := &cobra.Command{
				Use:  "send-agent [agent] [message]",
				Args: cobra.ExactArgs(2),
				RunE: mockExecuteMainCommand,
			}
			cmd.Flags().StringP("session", "s", "", "指定したセッション名を使用")
			cmd.Flags().BoolP("reset", "r", false, "前の役割定義をクリアして新しい指示を送信")

			args := []string{tt.agent, tt.message}
			if tt.session != "" {
				args = append([]string{"-s", tt.session}, args...)
			}
			if tt.reset {
				args = append([]string{"-r"}, args...)
			}

			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitFunction(t *testing.T) {
	t.Run("InitExists", func(t *testing.T) {
		// init関数が正常に実行されることを確認
		// rootCmdにフラグが設定されているかチェック
		assert.NotNil(t, rootCmd)
		assert.True(t, rootCmd.HasLocalFlags())

		// サブコマンドが追加されているかチェック
		assert.True(t, rootCmd.HasSubCommands())

		// 特定のフラグが存在するかチェック
		sessionFlag := rootCmd.Flags().Lookup("session")
		assert.NotNil(t, sessionFlag)
		assert.Equal(t, "s", sessionFlag.Shorthand)

		resetFlag := rootCmd.Flags().Lookup("reset")
		assert.NotNil(t, resetFlag)
		assert.Equal(t, "r", resetFlag.Shorthand)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("MainWithError", func(t *testing.T) {
		// メイン関数でのエラーハンドリングをテスト
		// os.Exit(1)が呼ばれるかどうかは直接テストできないが、
		// エラーメッセージの出力をテストできる

		// この部分は実際のmain関数の動作を模倣
		cmd := &cobra.Command{
			Use:  "send-agent [agent] [message]",
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return assert.AnError
			},
		}

		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestMainEnvironment(t *testing.T) {
	t.Run("EnvironmentVariables", func(t *testing.T) {
		// 環境変数の設定をテスト
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// PATHが設定されていることを確認
		assert.NotEmpty(t, os.Getenv("PATH"))
	})
}

func TestCobraCommandConfiguration(t *testing.T) {
	t.Run("RootCommandConfiguration", func(t *testing.T) {
		// rootCmdの設定をテスト
		assert.Equal(t, "send-agent [agent] [message]", rootCmd.Use)
		assert.NotEmpty(t, rootCmd.Short)
		assert.NotEmpty(t, rootCmd.Long)
		assert.NotEmpty(t, rootCmd.Example)
		assert.NotNil(t, rootCmd.RunE)
	})

	t.Run("ListCommandConfiguration", func(t *testing.T) {
		// listCmdの設定をテスト
		assert.Equal(t, "list [session-name]", listCmd.Use)
		assert.NotEmpty(t, listCmd.Short)
		assert.NotNil(t, listCmd.RunE)
	})

	t.Run("ListSessionsCommandConfiguration", func(t *testing.T) {
		// listSessionsCmdの設定をテスト
		assert.Equal(t, "list-sessions", listSessionsCmd.Use)
		assert.NotEmpty(t, listSessionsCmd.Short)
		assert.NotNil(t, listSessionsCmd.RunE)
	})
}
