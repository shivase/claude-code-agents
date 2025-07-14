package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// DirectoryResolver ディレクトリ解決機能
type DirectoryResolver struct {
	originalWorkingDir string
	projectRoot        string
	binaryPath         string
	isInSubdirectory   bool
}

// NewDirectoryResolver 新しいディレクトリ解決器を作成
func NewDirectoryResolver() *DirectoryResolver {
	resolver := &DirectoryResolver{}
	if err := resolver.Initialize(); err != nil {
		// エラーログは記録するが、処理を続行
		fmt.Fprintf(os.Stderr, "Warning: directory resolver initialization failed: %v\n", err)
	}
	return resolver
}

// Initialize ディレクトリ解決器の初期化
func (dr *DirectoryResolver) Initialize() error {
	// 元の作業ディレクトリを保存
	if wd, err := os.Getwd(); err == nil {
		dr.originalWorkingDir = wd
	} else {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// バイナリのパスを取得
	if exe, err := os.Executable(); err == nil {
		dr.binaryPath = exe
	} else {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// プロジェクトルートを決定
	dr.determineProjectRoot()

	// サブディレクトリ内実行かどうかを判定
	dr.isInSubdirectory = dr.isRunningFromSubdirectory()

	log.Info().
		Str("original_working_dir", dr.originalWorkingDir).
		Str("project_root", dr.projectRoot).
		Str("binary_path", dr.binaryPath).
		Bool("is_in_subdirectory", dr.isInSubdirectory).
		Msg("Directory resolver initialized")

	return nil
}

// determineProjectRoot プロジェクトルートを決定
func (dr *DirectoryResolver) determineProjectRoot() {
	// 1. バイナリのパスから推測
	binaryDir := filepath.Dir(dr.binaryPath)

	// バイナリがbuildディレクトリにある場合、その親ディレクトリを確認
	if strings.HasSuffix(binaryDir, "build") {
		parentDir := filepath.Dir(binaryDir)
		if dr.isProjectRoot(parentDir) {
			dr.projectRoot = parentDir
			return
		}
	}

	// 2. 現在のディレクトリから上位に向かって検索
	searchDir := dr.originalWorkingDir
	for {
		if dr.isProjectRoot(searchDir) {
			dr.projectRoot = searchDir
			return
		}

		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			// ルートディレクトリに到達
			break
		}
		searchDir = parent
	}

	// 3. デフォルトとして現在のディレクトリを使用
	dr.projectRoot = dr.originalWorkingDir
}

// isProjectRoot プロジェクトルートかどうかを判定
func (dr *DirectoryResolver) isProjectRoot(dir string) bool {
	// プロジェクトルートの特徴的なファイル・ディレクトリをチェック
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

// isRunningFromSubdirectory サブディレクトリから実行されているかを判定
func (dr *DirectoryResolver) isRunningFromSubdirectory() bool {
	// 現在のディレクトリがプロジェクトルートではない場合
	return dr.originalWorkingDir != dr.projectRoot
}

// GetOptimalWorkingDirectory 最適な作業ディレクトリを取得
func (dr *DirectoryResolver) GetOptimalWorkingDirectory() string {
	// start-agentsディレクトリから実行されている場合は、プロジェクトルートを返す
	if dr.isInSubdirectory {
		log.Info().
			Str("original", dr.originalWorkingDir).
			Str("optimal", dr.projectRoot).
			Msg("Using project root as working directory due to subdirectory execution")
		return dr.projectRoot
	}

	// それ以外は元の作業ディレクトリを返す
	return dr.originalWorkingDir
}

// GetProjectRoot プロジェクトルートを取得
func (dr *DirectoryResolver) GetProjectRoot() string {
	return dr.projectRoot
}

// GetOriginalWorkingDirectory 元の作業ディレクトリを取得
func (dr *DirectoryResolver) GetOriginalWorkingDirectory() string {
	return dr.originalWorkingDir
}

// IsInSubdirectory サブディレクトリから実行されているかを確認
func (dr *DirectoryResolver) IsInSubdirectory() bool {
	return dr.isInSubdirectory
}

// ResolveRelativePath 相対パスを適切に解決
func (dr *DirectoryResolver) ResolveRelativePath(relativePath string) string {
	// チルダ展開を先に実行
	expandedPath := ExpandPathSafe(relativePath)

	// 絶対パスの場合はセキュリティ検証を行う
	if filepath.IsAbs(expandedPath) {
		// 危険なシステムパスへのアクセスを防止
		if dr.isDangerousPath(expandedPath) {
			log.Warn().Str("path", expandedPath).Msg("Dangerous path access blocked")
			// 安全なプロジェクトルート内のパスに変更
			safePath := filepath.Join(dr.projectRoot, filepath.Base(expandedPath))
			return safePath
		}
		return expandedPath
	}

	// 相対パスの場合、プロジェクトルートを基準に解決
	resolved := filepath.Join(dr.projectRoot, expandedPath)

	// パストラバーサル攻撃の検証
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

// isDangerousPath 危険なシステムパスかどうかを判定
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

// isPathTraversal パストラバーサル攻撃かどうかを判定
func (dr *DirectoryResolver) isPathTraversal(resolvedPath string) bool {
	// プロジェクトルートの外に出ようとしているかチェック
	relPath, err := filepath.Rel(dr.projectRoot, resolvedPath)
	if err != nil {
		return true // エラーが発生した場合は危険とみなす
	}

	// "../"で始まる場合はパストラバーサル
	return strings.HasPrefix(relPath, "..")
}

// EnsureDirectoryExists ディレクトリの存在を確認し、必要に応じて作成
func (dr *DirectoryResolver) EnsureDirectoryExists(path string) error {
	// パスを解決（チルダ展開含む）
	resolvedPath := dr.ResolveRelativePath(path)

	// ディレクトリが存在するかチェック
	_, err := os.Stat(resolvedPath)
	switch {
	case os.IsNotExist(err):
		// ディレクトリを作成
		if err := os.MkdirAll(resolvedPath, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", resolvedPath, err)
		}
		log.Info().Str("directory", resolvedPath).Msg("Created directory")
	case err != nil:
		return fmt.Errorf("failed to check directory %s: %w", resolvedPath, err)
	}

	return nil
}

// GetRelativePathFromRoot プロジェクトルートからの相対パスを取得
func (dr *DirectoryResolver) GetRelativePathFromRoot(absolutePath string) string {
	if relPath, err := filepath.Rel(dr.projectRoot, absolutePath); err == nil {
		return relPath
	}
	return absolutePath
}

// ValidateWorkingDirectory 作業ディレクトリの有効性を検証
func (dr *DirectoryResolver) ValidateWorkingDirectory(workingDir string) error {
	// ディレクトリが存在するかチェック
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", workingDir)
	}

	// ディレクトリにアクセス可能かチェック
	if err := os.Chdir(workingDir); err != nil {
		return fmt.Errorf("cannot access working directory: %s", workingDir)
	}

	// 元のディレクトリに戻る
	if err := os.Chdir(dr.originalWorkingDir); err != nil {
		log.Warn().Err(err).Msg("Failed to return to original working directory")
	}

	return nil
}

// GetDirectoryInfo ディレクトリ情報を取得
func (dr *DirectoryResolver) GetDirectoryInfo() map[string]string {
	return map[string]string{
		"original_working_dir": dr.originalWorkingDir,
		"project_root":         dr.projectRoot,
		"binary_path":          dr.binaryPath,
		"optimal_working_dir":  dr.GetOptimalWorkingDirectory(),
		"is_in_subdirectory":   fmt.Sprintf("%t", dr.isInSubdirectory),
	}
}

// ConfigInterface インターフェイス定義
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

// FixDirectoryDependentPaths ディレクトリ依存パスの修正
func (dr *DirectoryResolver) FixDirectoryDependentPaths(config ConfigInterface) error {
	// 作業ディレクトリを最適化
	config.SetWorkingDir(dr.GetOptimalWorkingDirectory())

	// 相対パスを絶対パスに変換
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

// DisplayDirectoryInfo ディレクトリ情報を表示
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

// グローバルディレクトリ解決器
var globalDirectoryResolver *DirectoryResolver

// GetGlobalDirectoryResolver グローバルディレクトリ解決器を取得
func GetGlobalDirectoryResolver() *DirectoryResolver {
	if globalDirectoryResolver == nil {
		globalDirectoryResolver = NewDirectoryResolver()
	}
	return globalDirectoryResolver
}

// InitializeDirectoryResolver ディレクトリ解決器の初期化
func InitializeDirectoryResolver() error {
	resolver := GetGlobalDirectoryResolver()
	return resolver.Initialize()
}
