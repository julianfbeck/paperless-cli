package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the CLI configuration
type Config struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// configDir returns the config directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "paperless-cli"), nil
}

// configPath returns the config file path
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load loads the configuration from file
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetURL returns the Paperless URL from env or config
func GetURL() string {
	if url := os.Getenv("PAPERLESS_URL"); url != "" {
		return url
	}
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.URL
}

// GetToken returns the API token from env or config
func GetToken() string {
	if token := os.Getenv("PAPERLESS_TOKEN"); token != "" {
		return token
	}
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.Token
}

// SetURL saves the URL to config
func SetURL(url string) error {
	cfg, err := Load()
	if err != nil {
		cfg = &Config{}
	}
	cfg.URL = url
	return Save(cfg)
}

// SetToken saves the token to config
func SetToken(token string) error {
	cfg, err := Load()
	if err != nil {
		cfg = &Config{}
	}
	cfg.Token = token
	return Save(cfg)
}
