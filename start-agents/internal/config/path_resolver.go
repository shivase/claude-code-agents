package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathResolverInterface パス解決インターフェース
type PathResolverInterface interface {
	// ResolvePath パスを解決（環境変数展開、相対パス解決）
	ResolvePath(path string) (string, error)

	// ExpandEnvironmentVariables 環境変数を展開
	ExpandEnvironmentVariables(path string) string

	// ResolveTildePath チルダパス（~）を展開
	ResolveTildePath(path string) (string, error)

	// MakeAbsolutePath 絶対パスに変換
	MakeAbsolutePath(path, basePath string) (string, error)
}

// PathResolver パス解決器
type PathResolver struct {
	baseDir string
}

// NewPathResolver 新しいパス解決器を作成
func NewPathResolver(baseDir string) *PathResolver {
	return &PathResolver{
		baseDir: baseDir,
	}
}

// ResolvePath パスを解決
func (pr *PathResolver) ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	// 1. 環境変数の展開
	expanded := pr.ExpandEnvironmentVariables(path)

	// 2. チルダ展開
	tildeExpanded, err := pr.ResolveTildePath(expanded)
	if err != nil {
		return "", fmt.Errorf("tilde expansion failed: %w", err)
	}

	// 3. 絶対パス変換
	absolutePath, err := pr.MakeAbsolutePath(tildeExpanded, pr.baseDir)
	if err != nil {
		return "", fmt.Errorf("absolute path conversion failed: %w", err)
	}

	// 4. パスの正規化
	return filepath.Clean(absolutePath), nil
}

// ExpandEnvironmentVariables 環境変数を展開
func (pr *PathResolver) ExpandEnvironmentVariables(path string) string {
	return os.ExpandEnv(path)
}

// ResolveTildePath チルダパス（~）を展開
func (pr *PathResolver) ResolveTildePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, path[2:]), nil
}

// MakeAbsolutePath 絶対パスに変換
func (pr *PathResolver) MakeAbsolutePath(path, basePath string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	if basePath == "" {
		// ベースパスが空の場合は現在のディレクトリを使用
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		basePath = cwd
	}

	return filepath.Join(basePath, path), nil
}
