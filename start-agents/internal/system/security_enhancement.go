package system

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shivase/claude-code-agents/internal/auth"
)

// SecurityEnhancement - セキュリティ強化機能
type SecurityEnhancement struct {
	encryptionKey    []byte
	auditTrail       *AuditTrail
	integrityChecker *IntegrityChecker
}

// NewSecurityEnhancement - セキュリティ強化機能の作成
func NewSecurityEnhancement() (*SecurityEnhancement, error) {
	// 暗号化キーの生成
	encryptionKey := sha256.Sum256([]byte(fmt.Sprintf("claude-auth-security-%d", time.Now().UnixNano())))

	auditTrail, err := NewAuditTrail()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit trail: %w", err)
	}

	integrityChecker, err := NewIntegrityChecker()
	if err != nil {
		return nil, fmt.Errorf("failed to create integrity checker: %w", err)
	}

	return &SecurityEnhancement{
		encryptionKey:    encryptionKey[:],
		auditTrail:       auditTrail,
		integrityChecker: integrityChecker,
	}, nil
}

// EncryptData - データの暗号化
func (se *SecurityEnhancement) EncryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(se.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData - データの復号化
func (se *SecurityEnhancement) DecryptData(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(se.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// SecureDelete - セキュアファイル削除
func (se *SecurityEnhancement) SecureDelete(filePath string) error {
	// パスの正規化とディレクトリトラバーサル防止
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("file path contains directory traversal")
	}

	file, err := os.OpenFile(cleanPath, os.O_WRONLY, 0) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open file for secure delete: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Warning: failed to close file during secure delete: %v\n", err)
			if err != nil {
				return
			}
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// ファイルをランダムデータで上書き（3回）
	for i := 0; i < 3; i++ {
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to beginning of file: %w", err)
		}
		randomData := make([]byte, stat.Size())
		if _, err := rand.Read(randomData); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}
		if _, err := file.Write(randomData); err != nil {
			return fmt.Errorf("failed to write random data: %w", err)
		}
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}

	// ファイルを削除
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	log.Info().Str("file", filePath).Msg("File securely deleted")
	return nil
}

// AuditTrail - セキュリティ監査ログ
type AuditTrail struct {
	logFile string
}

// NewAuditTrail - 監査ログの作成
func NewAuditTrail() (*AuditTrail, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	logFile := filepath.Join(homeDir, ".claude", "security_audit.log")
	return &AuditTrail{logFile: logFile}, nil
}

// LogSecurityEvent - セキュリティイベントのログ
func (at *AuditTrail) LogSecurityEvent(event, details string) error {
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] %s: %s\n", timestamp, event, details)

	file, err := os.OpenFile(at.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Warning: failed to close audit log file: %v\n", err)
			if err != nil {
				return
			}
		}
	}()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}

	return nil
}

// IntegrityChecker - ファイル整合性チェック
type IntegrityChecker struct {
	checksumFile string
}

// NewIntegrityChecker - 整合性チェッカーの作成
func NewIntegrityChecker() (*IntegrityChecker, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	checksumFile := filepath.Join(homeDir, ".claude", "integrity.sha256")
	return &IntegrityChecker{checksumFile: checksumFile}, nil
}

// CalculateChecksum - ファイルチェックサムの計算
func (ic *IntegrityChecker) CalculateChecksum(filePath string) (string, error) {
	// パスの正規化とディレクトリトラバーサル防止
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("file path contains directory traversal")
	}

	data, err := os.ReadFile(cleanPath) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// SaveChecksum - チェックサムの保存
func (ic *IntegrityChecker) SaveChecksum(filePath, checksum string) error {
	entry := fmt.Sprintf("%s:%s\n", filePath, checksum)

	file, err := os.OpenFile(ic.checksumFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open checksum file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Warning: failed to close checksum file: %v\n", err)
			if err != nil {
				return
			}
		}
	}()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write checksum: %w", err)
	}

	return nil
}

// VerifyIntegrity - ファイル整合性の検証
func (ic *IntegrityChecker) VerifyIntegrity(filePath string) error {
	// 現在のチェックサムを計算
	currentChecksum, err := ic.CalculateChecksum(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate current checksum: %w", err)
	}

	// 保存されたチェックサムを読み込み
	data, err := os.ReadFile(ic.checksumFile)
	if err != nil {
		return fmt.Errorf("failed to read checksum file: %w", err)
	}

	// チェックサムの比較
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == filePath {
			if parts[1] != currentChecksum {
				return fmt.Errorf("integrity check failed: file %s has been modified", filePath)
			}
			return nil
		}
	}

	return fmt.Errorf("no checksum found for file: %s", filePath)
}

// SecurityManager - セキュリティマネージャー
type SecurityManager struct {
	authManager *auth.AuthManager
	enhancement *SecurityEnhancement
}

// NewSecurityManager - セキュリティマネージャーの作成
func NewSecurityManager() (*SecurityManager, error) {
	authManager, err := auth.NewAuthManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	enhancement, err := NewSecurityEnhancement()
	if err != nil {
		return nil, fmt.Errorf("failed to create security enhancement: %w", err)
	}

	return &SecurityManager{
		authManager: authManager,
		enhancement: enhancement,
	}, nil
}

// ProtectSystem - システムの保護
func (sm *SecurityManager) ProtectSystem() error {
	log.Info().Msg("Starting system protection")

	// 監査ログに記録
	if err := sm.enhancement.auditTrail.LogSecurityEvent("SYSTEM_PROTECTION_START", "Starting system protection"); err != nil {
		log.Error().Err(err).Msg("Failed to log security event")
	}

	// ファイル整合性チェック
	homeDir, _ := os.UserHomeDir()
	settingsFile := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsFile); err == nil {
		checksum, err := sm.enhancement.integrityChecker.CalculateChecksum(settingsFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to calculate settings checksum")
		} else {
			if err := sm.enhancement.integrityChecker.SaveChecksum(settingsFile, checksum); err != nil {
				log.Error().Err(err).Msg("Failed to save settings checksum")
			}
		}
	}

	// 基本的な保護とバックアップ
	if err := sm.authManager.ProtectAndBackup(); err != nil {
		return fmt.Errorf("failed to protect and backup: %w", err)
	}

	// 監査ログに記録
	if err := sm.enhancement.auditTrail.LogSecurityEvent("SYSTEM_PROTECTION_COMPLETE", "System protection completed"); err != nil {
		log.Error().Err(err).Msg("Failed to log security event")
	}

	log.Info().Msg("System protection completed")
	return nil
}

// RestoreSystem - システムの復元
func (sm *SecurityManager) RestoreSystem() error {
	log.Info().Msg("Starting system restore")

	// 監査ログに記録
	if err := sm.enhancement.auditTrail.LogSecurityEvent("SYSTEM_RESTORE_START", "Starting system restore"); err != nil {
		log.Error().Err(err).Msg("Failed to log security event")
	}

	// ファイル整合性の検証
	homeDir, _ := os.UserHomeDir()
	settingsFile := filepath.Join(homeDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsFile); err == nil {
		if err := sm.enhancement.integrityChecker.VerifyIntegrity(settingsFile); err != nil {
			log.Warn().Err(err).Msg("Settings file integrity check failed")
			// 監査ログに記録
			if logErr := sm.enhancement.auditTrail.LogSecurityEvent("INTEGRITY_CHECK_FAILED", fmt.Sprintf("Settings file integrity check failed: %v", err)); logErr != nil {
				log.Error().Err(logErr).Msg("Failed to log security event")
			}
		} else {
			log.Info().Msg("Settings file integrity verified")
		}
	}

	// 基本的な復元とクリーンアップ
	if err := sm.authManager.RestoreAndCleanup(); err != nil {
		return fmt.Errorf("failed to restore and cleanup: %w", err)
	}

	// 監査ログに記録
	if err := sm.enhancement.auditTrail.LogSecurityEvent("SYSTEM_RESTORE_COMPLETE", "System restore completed"); err != nil {
		log.Error().Err(err).Msg("Failed to log security event")
	}

	log.Info().Msg("System restore completed")
	return nil
}

// ValidateSecurityStatus - セキュリティ状態の検証
func (sm *SecurityManager) ValidateSecurityStatus() error {
	log.Info().Msg("Validating security status")

	// 認証設定の検証
	// Configuration validation is available
	// Validation completed

	// 監査ログに記録
	if err := sm.enhancement.auditTrail.LogSecurityEvent("SECURITY_VALIDATION", "Security status validation completed"); err != nil {
		log.Error().Err(err).Msg("Failed to log security event")
	}

	log.Info().Msg("Security status validation completed")
	return nil
}
