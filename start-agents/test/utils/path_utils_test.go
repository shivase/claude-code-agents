package utils

import (
	"github.com/shivase/claude-code-agents/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected string
		wantErr  bool
	}{
		{
			name:     "Empty string",
			path:     "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Tilde only",
			path:     "~",
			expected: homeDir,
			wantErr:  false,
		},
		{
			name:     "Tilde and relative path",
			path:     "~/documents",
			expected: filepath.Join(homeDir, "documents"),
			wantErr:  false,
		},
		{
			name:     "Tilde and slash",
			path:     "~/",
			expected: homeDir,
			wantErr:  false,
		},
		{
			name:     "User-specified path (not expanded)",
			path:     "~otheruser/documents",
			expected: "~otheruser/documents",
			wantErr:  false,
		},
		{
			name:     "Normal absolute path",
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
			wantErr:  false,
		},
		{
			name:     "Normal relative path",
			path:     "relative/path",
			expected: "relative/path",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ExpandPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExpandPathSafe(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Normal tilde path",
			path:     "~/test",
			expected: filepath.Join(homeDir, "test"),
		},
		{
			name:     "Normal path",
			path:     "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "Empty string",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExpandPathSafe(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	workDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name:    "Empty string",
			path:    "",
			wantErr: false,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "", result)
			},
		},
		{
			name:    "Tilde path",
			path:    "~/test",
			wantErr: false,
			check: func(t *testing.T, result string) {
				expected := filepath.Join(homeDir, "test")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:    "Relative path",
			path:    "relative/path",
			wantErr: false,
			check: func(t *testing.T, result string) {
				expected := filepath.Join(workDir, "relative/path")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:    "Absolute path",
			path:    "/tmp/test",
			wantErr: false,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "/tmp/test", result)
			},
		},
		{
			name:    "Current directory",
			path:    ".",
			wantErr: false,
			check: func(t *testing.T, result string) {
				assert.Equal(t, workDir, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.NormalizePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.check(t, result)
			}
		})
	}
}

func TestEnsureDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Create new directory",
			path:    filepath.Join(tmpDir, "newdir"),
			wantErr: false,
		},
		{
			name:    "Existing directory",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "Create nested directory",
			path:    filepath.Join(tmpDir, "nested/deep/dir"),
			wantErr: false,
		},
		{
			name:    "Tilde path",
			path:    "~/test_ensure_dir",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.EnsureDirectory(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Confirm directory was actually created
				normalizedPath, _ := utils.NormalizePath(tt.path)
				info, err := os.Stat(normalizedPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			}
		})
	}

	// Cleanup tilde path
	t.Cleanup(func() {
		homeDir, _ := os.UserHomeDir()
		testDir := filepath.Join(homeDir, "test_ensure_dir")
		_ = os.RemoveAll(testDir)
	})
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing file",
			path:     existingFile,
			expected: true,
		},
		{
			name:     "Existing directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "Non-existing file",
			path:     filepath.Join(tmpDir, "nonexistent.txt"),
			expected: false,
		},
		{
			name:     "Non-existing directory",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
		{
			name:     "Empty string",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.PathExists(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "File",
			path:     testFile,
			expected: false,
		},
		{
			name:     "Non-existing path",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
		{
			name:     "Empty string",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsDirectory(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		base     string
		elements []string
		wantErr  bool
		check    func(t *testing.T, result string)
	}{
		{
			name:     "Normal path join",
			base:     "/usr/local",
			elements: []string{"bin", "claude"},
			wantErr:  false,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "/usr/local/bin/claude", result)
			},
		},
		{
			name:     "Tilde base path",
			base:     "~",
			elements: []string{"documents", "test.txt"},
			wantErr:  false,
			check: func(t *testing.T, result string) {
				expected := filepath.Join(homeDir, "documents", "test.txt")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:     "Relative path join",
			base:     "relative",
			elements: []string{"sub", "file.txt"},
			wantErr:  false,
			check: func(t *testing.T, result string) {
				workDir, _ := os.Getwd()
				expected := filepath.Join(workDir, "relative", "sub", "file.txt")
				assert.Equal(t, expected, result)
			},
		},
		{
			name:     "Empty elements",
			base:     "/tmp",
			elements: []string{},
			wantErr:  false,
			check: func(t *testing.T, result string) {
				assert.Equal(t, "/tmp", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.JoinPath(tt.base, tt.elements...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.check(t, result)
			}
		})
	}
}

func TestJoinPathSafe(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		elements []string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "Normal path join",
			base:     "/usr/local",
			elements: []string{"bin", "claude"},
			check: func(t *testing.T, result string) {
				assert.Equal(t, "/usr/local/bin/claude", result)
			},
		},
		{
			name:     "Tilde path",
			base:     "~/test",
			elements: []string{"sub"},
			check: func(t *testing.T, result string) {
				homeDir, _ := os.UserHomeDir()
				expected := filepath.Join(homeDir, "test", "sub")
				assert.Equal(t, expected, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.JoinPathSafe(tt.base, tt.elements...)
			tt.check(t, result)
		})
	}
}

// Benchmark tests
func BenchmarkExpandPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = utils.ExpandPath("~/test/path")
	}
}

func BenchmarkExpandPathSafe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.ExpandPathSafe("~/test/path")
	}
}

func BenchmarkNormalizePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = utils.NormalizePath("~/test/path")
	}
}
