package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Logging  LoggingConfig  `yaml:"logging"`
	Card     CardConfig     `yaml:"card"`
	Cache    CacheConfig    `yaml:"cache"`
	Security SecurityConfig `yaml:"security"`
	Advanced AdvancedConfig `yaml:"advanced"`
	PKCS11   PKCS11Config   `yaml:"pkcs11"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
}

// CardConfig holds card reader configuration
type CardConfig struct {
	TimeoutSeconds int `yaml:"timeout_seconds"`
	RetryAttempts  int `yaml:"retry_attempts"`
	RetryDelayMS   int `yaml:"retry_delay_ms"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled         bool `yaml:"enabled"`
	DurationSeconds int  `yaml:"duration_seconds"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	AllowedExtensions []string `yaml:"allowed_extensions"`
	StrictValidation  bool     `yaml:"strict_validation"`
}

// AdvancedConfig holds advanced configuration
type AdvancedConfig struct {
	VerboseAPDU bool              `yaml:"verbose_apdu"`
	CustomFIDs  map[string]string `yaml:"custom_fids"`
}

// PKCS11Config holds defaults for local token decryption.
type PKCS11Config struct {
	LibraryPath string `yaml:"library_path"`
	Slot        int    `yaml:"slot"`
	Label       string `yaml:"label"`
	KeyIDHex    string `yaml:"key_id_hex"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Logging: LoggingConfig{
			Enabled:    true,
			Level:      "info",
			File:       "",
			MaxSizeMB:  10,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   true,
		},
		Card: CardConfig{
			TimeoutSeconds: 30,
			RetryAttempts:  3,
			RetryDelayMS:   500,
		},
		Cache: CacheConfig{
			Enabled:         true,
			DurationSeconds: 300,
		},
		Security: SecurityConfig{
			AllowedExtensions: []string{},
			StrictValidation:  true,
		},
		Advanced: AdvancedConfig{
			VerboseAPDU: false,
			CustomFIDs:  make(map[string]string),
		},
		PKCS11: PKCS11Config{
			LibraryPath: "",
			Slot:        0,
			Label:       "",
			KeyIDHex:    "",
		},
	}
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Start with defaults
	cfg := Default()

	// If path is empty, try to find config.yaml in common locations
	if path == "" {
		path = findConfigFile()
	}

	// If still no path, return defaults
	if path == "" {
		return cfg, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// findConfigFile searches for config.yaml in common locations
func findConfigFile() string {
	// Possible locations (in order of preference)
	locations := []string{
		"config.yaml",   // Current directory
		"./config.yaml", // Explicit current directory
		filepath.Join(getExecutableDir(), "config.yaml"), // Same directory as executable
	}

	// Platform-specific locations
	if configDir := getConfigDir(); configDir != "" {
		locations = append(locations, filepath.Join(configDir, "eid-pd-reader", "config.yaml"))
	}

	// Check each location
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// getConfigDir returns the platform-specific config directory
func getConfigDir() string {
	// Windows: %PROGRAMDATA%\eid-pd-reader
	// Linux: /etc/eid-pd-reader
	// macOS: /Library/Application Support/eid-pd-reader

	if dir := os.Getenv("PROGRAMDATA"); dir != "" {
		// Windows
		return dir
	}

	// Unix-like systems
	if _, err := os.Stat("/etc"); err == nil {
		return "/etc"
	}

	return ""
}
