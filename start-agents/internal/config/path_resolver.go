package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathResolverInterface defines path resolver interface
type PathResolverInterface interface {
	// ResolvePath resolves path (environment variable expansion, relative path resolution)
	ResolvePath(path string) (string, error)

	// ExpandEnvironmentVariables expands environment variables
	ExpandEnvironmentVariables(path string) string

	// ResolveTildePath expands tilde path (~)
	ResolveTildePath(path string) (string, error)

	// MakeAbsolutePath converts to absolute path
	MakeAbsolutePath(path, basePath string) (string, error)
}

// PathResolver resolves file paths
type PathResolver struct {
	baseDir string
}

// NewPathResolver creates a new path resolver
func NewPathResolver(baseDir string) *PathResolver {
	return &PathResolver{
		baseDir: baseDir,
	}
}

// ResolvePath resolves path
func (pr *PathResolver) ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	// 1. Environment variable expansion
	expanded := pr.ExpandEnvironmentVariables(path)

	// 2. Tilde expansion
	tildeExpanded, err := pr.ResolveTildePath(expanded)
	if err != nil {
		return "", fmt.Errorf("tilde expansion failed: %w", err)
	}

	// 3. Absolute path conversion
	absolutePath, err := pr.MakeAbsolutePath(tildeExpanded, pr.baseDir)
	if err != nil {
		return "", fmt.Errorf("absolute path conversion failed: %w", err)
	}

	// 4. Path normalization
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
		// Use current directory if base path is empty
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		basePath = cwd
	}

	return filepath.Join(basePath, path), nil
}
