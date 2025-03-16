package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Bastions map[string]*BastionConfig `yaml:"bastions"`
}

// BastionConfig holds the configuration for a single bastion server
type BastionConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Port     int    `yaml:"port"`
	AuthType string `yaml:"auth_type"` // "key" or "password"
	KeyPath  string `yaml:"key_path,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// LoadConfig loads the configuration from the default location
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(home, ".mytunnel", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{Bastions: make(map[string]*BastionConfig)}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the default location
func SaveConfig(config *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(home, ".mytunnel")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddBastion adds a new bastion configuration
func (c *Config) AddBastion(name string, bastion *BastionConfig) {
	if c.Bastions == nil {
		c.Bastions = make(map[string]*BastionConfig)
	}
	c.Bastions[name] = bastion
}

// RemoveBastion removes a bastion configuration
func (c *Config) RemoveBastion(name string) {
	delete(c.Bastions, name)
}

// GetBastion retrieves a bastion configuration by name
func (c *Config) GetBastion(name string) (*BastionConfig, bool) {
	bastion, ok := c.Bastions[name]
	return bastion, ok
} 