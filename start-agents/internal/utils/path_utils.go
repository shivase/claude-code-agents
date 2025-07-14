package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath チルダ（~）を含むパスを展開する
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// チルダで始まる場合
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}

		// ~/foo -> /home/user/foo
		if len(path) == 1 || path[1] == '/' {
			return filepath.Join(homeDir, path[1:]), nil
		}

		// ~user/foo の場合は展開しない（ユーザー指定）
		return path, nil
	}

	return path, nil
}

// ExpandPathSafe チルダ展開のセーフ版（エラー時は元のパスを返す）
func ExpandPathSafe(path string) string {
	// URLエンコードされた攻撃的なパスを検出・拒否
	if strings.Contains(path, "%2F") || strings.Contains(path, "%2f") ||
		strings.Contains(path, "%5C") || strings.Contains(path, "%5c") ||
		strings.Contains(path, "%00") {
		// エンコードされた攻撃文字が含まれている場合は安全化
		path = strings.ReplaceAll(path, "%2F", "_")
		path = strings.ReplaceAll(path, "%2f", "_")
		path = strings.ReplaceAll(path, "%5C", "_")
		path = strings.ReplaceAll(path, "%5c", "_")
		path = strings.ReplaceAll(path, "%00", "")
	}

	expanded, err := ExpandPath(path)
	if err != nil {
		return path
	}
	return expanded
}

// NormalizePath パスの正規化（チルダ展開 + 絶対パス変換）
func NormalizePath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// チルダ展開
	expanded := ExpandPathSafe(path)

	// 絶対パス化
	if !filepath.IsAbs(expanded) {
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return expanded, err
		}
		return abs, nil
	}

	return filepath.Clean(expanded), nil
}

// EnsureDirectory ディレクトリの存在確認と作成
func EnsureDirectory(path string) error {
	normalizedPath, err := NormalizePath(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(normalizedPath, 0750)
}

// PathExists パスの存在確認
func PathExists(path string) bool {
	normalizedPath := ExpandPathSafe(path)
	_, err := os.Stat(normalizedPath)
	return err == nil
}

// IsDirectory ディレクトリかどうかを確認
func IsDirectory(path string) bool {
	normalizedPath := ExpandPathSafe(path)
	info, err := os.Stat(normalizedPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// JoinPath パスを結合し、チルダ展開する
func JoinPath(base string, elements ...string) (string, error) {
	path := filepath.Join(base, filepath.Join(elements...))
	return NormalizePath(path)
}

// JoinPathSafe パス結合のセーフ版
func JoinPathSafe(base string, elements ...string) string {
	path, err := JoinPath(base, elements...)
	if err != nil {
		return filepath.Join(base, filepath.Join(elements...))
	}
	return path
}
