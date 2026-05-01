package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration file.
type Config struct {
	CurrentDriver string              `json:"current_driver"`
	Drivers       map[string]*DriverConfig `json:"drivers"`
}

// DriverConfig holds configuration for a specific driver.
type DriverConfig struct {
	Type   string            `json:"type"`
	Tokens map[string]string `json:"tokens"`
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cloud-cli.json"
	}
	return filepath.Join(home, ".config", "cloud-cli", "config.json")
}

// LoadConfig reads the configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Drivers: make(map[string]*DriverConfig)}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Drivers == nil {
		cfg.Drivers = make(map[string]*DriverConfig)
	}
	return &cfg, nil
}

// SaveConfig writes the configuration to the given path.
func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// GetDriverConfig returns the configuration for a specific driver.
func (c *Config) GetDriverConfig(name string) (*DriverConfig, bool) {
	d, ok := c.Drivers[name]
	return d, ok
}

// SetDriverConfig sets the configuration for a specific driver.
func (c *Config) SetDriverConfig(name string, dc *DriverConfig) {
	if c.Drivers == nil {
		c.Drivers = make(map[string]*DriverConfig)
	}
	c.Drivers[name] = dc
	c.CurrentDriver = name
}
