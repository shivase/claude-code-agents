package auth

// AuthProviderInterface defines Claude authentication provider interface
type AuthProviderInterface interface {
	// CheckAuth checks authentication status
	CheckAuth() error
	// CheckSettingsFile checks settings file
	CheckSettingsFile() error
	// IsReady checks if Claude CLI is ready for use
	IsReady() bool
	// GetPath gets Claude CLI path
	GetPath() string
	// ValidateSetup performs comprehensive setup validation
	ValidateSetup() error
}

// PreAuthCheckerInterface defines pre-authentication checker interface
type PreAuthCheckerInterface interface {
	// CheckAuthenticationBeforeStart checks authentication before start
	CheckAuthenticationBeforeStart() error
}
