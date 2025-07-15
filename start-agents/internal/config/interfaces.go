package config

// ConfigLoaderInterface defines configuration loader interface
type ConfigLoaderInterface interface {
	// LoadTeamConfig loads team configuration
	LoadTeamConfig() (*TeamConfig, error)
	// SaveTeamConfig saves team configuration
	SaveTeamConfig(*TeamConfig) error
	// GetTeamConfigPath gets team configuration file path
	GetTeamConfigPath() string
}

// === Dynamic instruction feature already implemented ===
// Interfaces and structures are defined in instruction_resolver.go and instruction_validator.go

// ConfigGeneratorInterface defines configuration generator interface
type ConfigGeneratorInterface interface {
	// GenerateConfig generates configuration file
	GenerateConfig(forceOverwrite bool) error
	// ValidateConfig validates configuration file
	ValidateConfig() error
}
