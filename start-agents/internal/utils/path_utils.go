package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands paths containing tilde (~)
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// If starts with tilde
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}

		// ~/foo -> /home/user/foo
		if len(path) == 1 || path[1] == '/' {
			return filepath.Join(homeDir, path[1:]), nil
		}

		// Don't expand ~user/foo case (user specified)
		return path, nil
	}

	return path, nil
}

// ExpandPathSafe safe version of tilde expansion (returns original path on error)
func ExpandPathSafe(path string) string {
	// Detect and reject URL-encoded malicious paths
	if strings.Contains(path, "%2F") || strings.Contains(path, "%2f") ||
		strings.Contains(path, "%5C") || strings.Contains(path, "%5c") ||
		strings.Contains(path, "%00") {
		// Sanitize if encoded attack characters are found
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

// NormalizePath normalizes path (tilde expansion + absolute path conversion)
func NormalizePath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Tilde expansion
	expanded := ExpandPathSafe(path)

	// Make absolute path
	if !filepath.IsAbs(expanded) {
		abs, err := filepath.Abs(expanded)
		if err != nil {
			return expanded, err
		}
		return abs, nil
	}

	return filepath.Clean(expanded), nil
}

// EnsureDirectory checks directory existence and creates it
func EnsureDirectory(path string) error {
	normalizedPath, err := NormalizePath(path)
	if err != nil {
		return err
	}

	return os.MkdirAll(normalizedPath, 0750)
}

// PathExists checks path existence
func PathExists(path string) bool {
	normalizedPath := ExpandPathSafe(path)
	_, err := os.Stat(normalizedPath)
	return err == nil
}

// IsDirectory checks if path is directory
func IsDirectory(path string) bool {
	normalizedPath := ExpandPathSafe(path)
	info, err := os.Stat(normalizedPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// JoinPath joins paths and expands tilde
func JoinPath(base string, elements ...string) (string, error) {
	path := filepath.Join(base, filepath.Join(elements...))
	return NormalizePath(path)
}

// JoinPathSafe safe version of path joining
func JoinPathSafe(base string, elements ...string) string {
	path, err := JoinPath(base, elements...)
	if err != nil {
		return filepath.Join(base, filepath.Join(elements...))
	}
	return path
}
