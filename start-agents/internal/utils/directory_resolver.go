package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// DirectoryResolver resolves directory paths
type DirectoryResolver struct {
	originalWorkingDir string
	projectRoot        string
	binaryPath         string
	isInSubdirectory   bool
}

// NewDirectoryResolver creates a new directory resolver
func NewDirectoryResolver() *DirectoryResolver {
	resolver := &DirectoryResolver{}
	if err := resolver.Initialize(); err != nil {
		// Log error but continue processing
		fmt.Fprintf(os.Stderr, "Warning: directory resolver initialization failed: %v\n", err)
	}
	return resolver
}

// Initialize initializes the directory resolver
func (dr *DirectoryResolver) Initialize() error {
	// Save original working directory
	if wd, err := os.Getwd(); err == nil {
		dr.originalWorkingDir = wd
	} else {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Get binary path
	if exe, err := os.Executable(); err == nil {
		dr.binaryPath = exe
	} else {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Determine project root
	dr.determineProjectRoot()

	// Check if running from subdirectory
	dr.isInSubdirectory = dr.isRunningFromSubdirectory()

	log.Info().
		Str("original_working_dir", dr.originalWorkingDir).
		Str("project_root", dr.projectRoot).
		Str("binary_path", dr.binaryPath).
		Bool("is_in_subdirectory", dr.isInSubdirectory).
		Msg("Directory resolver initialized")

	return nil
}

// determineProjectRoot determines the project root
func (dr *DirectoryResolver) determineProjectRoot() {
	// 1. Infer from binary path
	binaryDir := filepath.Dir(dr.binaryPath)

	// If binary is in build directory, check its parent directory
	if strings.HasSuffix(binaryDir, "build") {
		parentDir := filepath.Dir(binaryDir)
		if dr.isProjectRoot(parentDir) {
			dr.projectRoot = parentDir
			return
		}
	}

	// 2. Search upward from current directory
	searchDir := dr.originalWorkingDir
	for {
		if dr.isProjectRoot(searchDir) {
			dr.projectRoot = searchDir
			return
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			// Reached root directory
			break
		}
		searchDir = parent
	}

	// 3. Use current directory as default
	dr.projectRoot = dr.originalWorkingDir
}

// isProjectRoot determines if directory is project root
func (dr *DirectoryResolver) isProjectRoot(dir string) bool {
	// Check for characteristic files/directories of project root
	indicators := []string{
		"start-agents",
		"send-agent",
		"docs",
		".git",
		"go.mod",
		"LICENSE",
	}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
			return true
		}
	}

	return false
}

// isRunningFromSubdirectory checks if running from subdirectory
func (dr *DirectoryResolver) isRunningFromSubdirectory() bool {
	// If current directory is not project root
	return dr.originalWorkingDir != dr.projectRoot
}

// GetOptimalWorkingDirectory gets the optimal working directory
func (dr *DirectoryResolver) GetOptimalWorkingDirectory() string {
	// Return project root if running from start-agents directory
	if dr.isInSubdirectory {
		log.Info().
			Str("original", dr.originalWorkingDir).
			Str("optimal", dr.projectRoot).
			Msg("Using project root as working directory due to subdirectory execution")
		return dr.projectRoot
	}

	// Otherwise return original working directory
	return dr.originalWorkingDir
}

// GetProjectRoot gets the project root
func (dr *DirectoryResolver) GetProjectRoot() string {
	return dr.projectRoot
}

// GetOriginalWorkingDirectory gets the original working directory
func (dr *DirectoryResolver) GetOriginalWorkingDirectory() string {
	return dr.originalWorkingDir
}

// IsInSubdirectory checks if running from subdirectory
func (dr *DirectoryResolver) IsInSubdirectory() bool {
	return dr.isInSubdirectory
}

// ResolveRelativePath resolves relative paths appropriately
func (dr *DirectoryResolver) ResolveRelativePath(relativePath string) string {
	// Perform tilde expansion first
	expandedPath := ExpandPathSafe(relativePath)

	// Perform security validation for absolute paths
	if filepath.IsAbs(expandedPath) {
		// Prevent access to dangerous system paths
		if dr.isDangerousPath(expandedPath) {
			log.Warn().Str("path", expandedPath).Msg("Dangerous path access blocked")
			// Change to safe path within project root
			safePath := filepath.Join(dr.projectRoot, filepath.Base(expandedPath))
			return safePath
		}
		return expandedPath
	}

	// For relative paths, resolve based on project root
	resolved := filepath.Join(dr.projectRoot, expandedPath)

	// Validate for path traversal attacks
	cleanResolved := filepath.Clean(resolved)
	if dr.isPathTraversal(cleanResolved) {
		log.Warn().
			Str("original_path", relativePath).
			Str("resolved_path", cleanResolved).
			Msg("Path traversal attack blocked")
		// 安全なプロジェクトルート内のパスに変更
		safePath := filepath.Join(dr.projectRoot, filepath.Base(relativePath))
		return safePath
	}

	log.Debug().
		Str("relative_path", relativePath).
		Str("expanded_path", expandedPath).
		Str("resolved_path", cleanResolved).
		Str("project_root", dr.projectRoot).
		Msg("Resolved relative path")

	return cleanResolved
}

// isDangerousPath determines if path is dangerous system path
func (dr *DirectoryResolver) isDangerousPath(path string) bool {
	dangerousPaths := []string{
		"/etc", "/root", "/home", "/usr/bin", "/usr/sbin",
		"/var", "/boot", "/dev", "/proc", "/sys", "/bin", "/sbin",
	}

	cleanPath := filepath.Clean(path)
	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(cleanPath, dangerous) {
			return true
		}
	}
	return false
}

// isPathTraversal determines if path is path traversal attack
func (dr *DirectoryResolver) isPathTraversal(resolvedPath string) bool {
	// Check if trying to go outside project root
	relPath, err := filepath.Rel(dr.projectRoot, resolvedPath)
	if err != nil {
		return true // Consider dangerous if error occurs
	}

	// Path traversal if starts with "../"
	return strings.HasPrefix(relPath, "..")
}

// EnsureDirectoryExists checks directory existence and creates if needed
func (dr *DirectoryResolver) EnsureDirectoryExists(path string) error {
	// Resolve path (including tilde expansion)
	resolvedPath := dr.ResolveRelativePath(path)

	// ディレクトリが存在するかチェック
	_, err := os.Stat(resolvedPath)
	switch {
	case os.IsNotExist(err):
		// Create directory
		if err := os.MkdirAll(resolvedPath, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", resolvedPath, err)
		}
		log.Info().Str("directory", resolvedPath).Msg("Created directory")
	case err != nil:
		return fmt.Errorf("failed to check directory %s: %w", resolvedPath, err)
	}

	return nil
}

// GetRelativePathFromRoot gets relative path from project root
func (dr *DirectoryResolver) GetRelativePathFromRoot(absolutePath string) string {
	if relPath, err := filepath.Rel(dr.projectRoot, absolutePath); err == nil {
		return relPath
	}
	return absolutePath
}

// ValidateWorkingDirectory validates working directory
func (dr *DirectoryResolver) ValidateWorkingDirectory(workingDir string) error {
	// ディレクトリが存在するかチェック
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", workingDir)
	}

	// Check if directory is accessible
	if err := os.Chdir(workingDir); err != nil {
		return fmt.Errorf("cannot access working directory: %s", workingDir)
	}

	// Return to original directory
	if err := os.Chdir(dr.originalWorkingDir); err != nil {
		log.Warn().Err(err).Msg("Failed to return to original working directory")
	}

	return nil
}

// GetDirectoryInfo gets directory information
func (dr *DirectoryResolver) GetDirectoryInfo() map[string]string {
	return map[string]string{
		"original_working_dir": dr.originalWorkingDir,
		"project_root":         dr.projectRoot,
		"binary_path":          dr.binaryPath,
		"optimal_working_dir":  dr.GetOptimalWorkingDirectory(),
		"is_in_subdirectory":   fmt.Sprintf("%t", dr.isInSubdirectory),
	}
}

// ConfigInterface defines configuration interface
type ConfigInterface interface {
	GetWorkingDir() string
	SetWorkingDir(string)
	GetClaudeCLIPath() string
	SetClaudeCLIPath(string)
	GetInstructionsDir() string
	SetInstructionsDir(string)
	GetConfigDir() string
	SetConfigDir(string)
	GetLogFile() string
	SetLogFile(string)
	GetAuthBackupDir() string
	SetAuthBackupDir(string)
}

// FixDirectoryDependentPaths fixes directory dependent paths
func (dr *DirectoryResolver) FixDirectoryDependentPaths(config ConfigInterface) error {
	// Optimize working directory
	config.SetWorkingDir(dr.GetOptimalWorkingDirectory())

	// Convert relative paths to absolute paths
	if !filepath.IsAbs(config.GetClaudeCLIPath()) {
		config.SetClaudeCLIPath(dr.ResolveRelativePath(config.GetClaudeCLIPath()))
	}

	if !filepath.IsAbs(config.GetInstructionsDir()) {
		config.SetInstructionsDir(dr.ResolveRelativePath(config.GetInstructionsDir()))
	}

	if !filepath.IsAbs(config.GetConfigDir()) {
		config.SetConfigDir(dr.ResolveRelativePath(config.GetConfigDir()))
	}

	if !filepath.IsAbs(config.GetLogFile()) {
		config.SetLogFile(dr.ResolveRelativePath(config.GetLogFile()))
	}

	if !filepath.IsAbs(config.GetAuthBackupDir()) {
		config.SetAuthBackupDir(dr.ResolveRelativePath(config.GetAuthBackupDir()))
	}

	return nil
}

// DisplayDirectoryInfo displays directory information
func (dr *DirectoryResolver) DisplayDirectoryInfo() {
	fmt.Println("\n📁 Directory Resolution Information")
	fmt.Println("==================================")

	info := dr.GetDirectoryInfo()
	fmt.Printf("   Original Working Dir: %s\n", info["original_working_dir"])
	fmt.Printf("   Project Root: %s\n", info["project_root"])
	fmt.Printf("   Binary Path: %s\n", info["binary_path"])
	fmt.Printf("   Optimal Working Dir: %s\n", info["optimal_working_dir"])
	fmt.Printf("   Is In Subdirectory: %s\n", info["is_in_subdirectory"])

	if dr.isInSubdirectory {
		fmt.Println("   ⚠️  Subdirectory execution detected - using project root")
	} else {
		fmt.Println("   ✅ Normal execution from project root")
	}

	fmt.Println()
}

// Global directory resolver
var globalDirectoryResolver *DirectoryResolver

// GetGlobalDirectoryResolver gets global directory resolver
func GetGlobalDirectoryResolver() *DirectoryResolver {
	if globalDirectoryResolver == nil {
		globalDirectoryResolver = NewDirectoryResolver()
	}
	return globalDirectoryResolver
}

// InitializeDirectoryResolver initializes directory resolver
func InitializeDirectoryResolver() error {
	resolver := GetGlobalDirectoryResolver()
	return resolver.Initialize()
}
