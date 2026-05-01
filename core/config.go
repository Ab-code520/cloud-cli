package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config represents the global configuration for cloud-cli.
type Config struct {
	mu       sync.RWMutex
	path     string
	Default  string            `yaml:"default"`
	Accounts map[string]*Account `yaml:"accounts"`
	Settings Settings           `yaml:"settings"`
}

// Account represents a single cloud drive account configuration.
type Account struct {
	Type   string            `yaml:"type"`
	Cookie map[string]string `yaml:"cookie"`
	Params map[string]string `yaml:"params,omitempty"`
}

// Settings represents global CLI settings.
type Settings struct {
	SpeedLimit string `yaml:"speed_limit"`
	Retries    int    `yaml:"retries"`
	Timeout    int    `yaml:"timeout"`
}

// GlobalConfig is the singleton configuration instance.
var (
	GlobalConfig *Config
	configOnce   sync.Once
)

// LoadConfig loads the configuration from the given path or default location.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		path = filepath.Join(home, ".config", "cloud-cli", "config.yaml")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	cfg := &Config{
		path:     path,
		Accounts: make(map[string]*Account),
		Settings: Settings{
			SpeedLimit: "0", // 0 means no limit
			Retries:    3,
			Timeout:    60,
		},
	}

	// Try to read existing config
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if not exists
			cfg.Default = "default"
			cfg.Accounts["default"] = &Account{
				Type:   "quark",
				Cookie: make(map[string]string),
			}
			if err := cfg.Save(); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Accounts == nil {
		cfg.Accounts = make(map[string]*Account)
	}

	return cfg, nil
}

// Save saves the current configuration to disk.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	return os.Rename(tmpPath, c.path)
}

// GetAccount returns the configuration for a specific account name.
func (c *Config) GetAccount(name string) (*Account, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if name == "" {
		name = c.Default
	}

	acc, ok := c.Accounts[name]
	if !ok {
		return nil, fmt.Errorf("account '%s' not found", name)
	}

	return acc, nil
}

// AddAccount adds or updates an account configuration.
func (c *Config) AddAccount(name string, acc *Account) error {
	c.mu.Lock()
	c.Accounts[name] = acc
	c.mu.Unlock()

	return c.Save()
}

// RemoveAccount removes an account configuration.
func (c *Config) RemoveAccount(name string) error {
	c.mu.Lock()
	delete(c.Accounts, name)
	if c.Default == name && len(c.Accounts) > 0 {
		// Set a new default
		for k := range c.Accounts {
			c.Default = k
			break
		}
	}
	c.mu.Unlock()

	return c.Save()
}

// ListAccounts returns a list of all configured account names.
func (c *Config) ListAccounts() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.Accounts))
	for name := range c.Accounts {
		names = append(names, name)
	}
	return names
}

// GetDefaultAccount returns the default account.
func (c *Config) GetDefaultAccount() (*Account, error) {
	return c.GetAccount("")
}
