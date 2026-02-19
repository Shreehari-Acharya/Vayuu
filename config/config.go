package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	instance *Config
	once     sync.Once
)

// Load loads and validates configuration.
// In development mode, it loads from .env/environment variables.
// Otherwise it loads from ~/.vayuu/vayuuConfig.json and runs setup if missing.
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		if isDevelopmentMode() {
			instance, err = loadFromEnv()
			return
		}

		instance, err = loadFromFileOrSetup()
	})

	if err != nil {
		return nil, err
	}
	return instance, nil
}

// validate checks that all required fields are present and valid
func (c *Config) validate() error {
	if c.TelegramToken == "" {
		return fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	if c.ApiKey == "" {
		return fmt.Errorf("API_KEY is required")
	}

	if c.ApiBaseURL == "" {
		return fmt.Errorf("API_BASE_URL is required")
	}

	if c.Model == "" {
		return fmt.Errorf("MODEL is required")
	}

	if c.AgentWorkDir == "" {
		return fmt.Errorf("AGENT_WORKDIR is required")
	}

	if c.AllowedUsername == "" {
		return fmt.Errorf("ALLOWED_USERNAME is required")
	}

	return nil
}

// loadFromEnv loads configuration from environment variables, typically used in development mode
func loadFromEnv() (*Config, error) {
	_ = godotenv.Load()

	cfg := configFromEnv(os.Getenv)
	if err := normalizeConfigPaths(cfg, false); err != nil {
		return nil, err
	}

	return cfg, cfg.validate()
}

// loadFromFileOrSetup loads the config from file or runs setup if the file doesn't exist
func loadFromFileOrSetup() (*Config, error) {
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to stat config file: %w", err)
		}

		if err := RunSetup(); err != nil {
			return nil, err
		}
	}

	if err := ensureConfigFilePermissions(configPath); err != nil {
		return nil, err
	}

	return loadFromFile(configPath)
}

// loadFromFile loads the config from the specified file path and validates it
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := normalizeConfigPaths(cfg, false); err != nil {
		return nil, err
	}

	return cfg, cfg.validate()
}

// saveToFile saves the config to the specified path with 0600 permissions
func saveToFile(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if err := ensureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	dir := filepath.Dir(path)
	tmpFile := filepath.Join(dir, ".vayuuConfig.tmp")
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, path); err != nil {
		return err
	}

	return ensureConfigFilePermissions(path)
}

// ensureConfigDir creates the config directory if it doesn't exist and sets permissions to 0700
func ensureConfigDir() error {
	configDir := filepath.Dir(getConfigPath())
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	return os.Chmod(configDir, 0700)
}

// ensureConfigFilePermissions sets file permissions to 0600 (user read/write only) for the config file
func ensureConfigFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Mode().Perm() != 0600 {
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("failed to set config permissions: %w", err)
		}
	}

	return nil
}

// getConfigPath returns the path to the config file, which is ~/.vayuu/vayuuConfig.json
func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".vayuu", configFileName)
}

// isDevelopmentMode checks environment variables to determine if the application is running in development mode
func isDevelopmentMode() bool {
	values := []string{
		os.Getenv("MODE"),
		os.Getenv("APP_ENV"),
		os.Getenv("ENV"),
		os.Getenv("VAYUU_ENV"),
	}

	for _, value := range values {
		if strings.EqualFold(value, "development") {
			return true
		}
	}

	return false
}

// configFromEnv constructs a Config struct from environment variables using the provided getEnv function (e.g., os.Getenv)
func configFromEnv(getEnv func(string) string) *Config {
	return &Config{
		TelegramToken:   getEnv("TELEGRAM_TOKEN"),
		ApiKey:          getEnv("API_KEY"),
		ApiBaseURL:      getEnv("API_BASE_URL"),
		Model:           getEnv("MODEL"),
		AgentWorkDir:    getEnv("AGENT_WORKDIR"),
		AllowedUsername: getEnv("ALLOWED_USERNAME"),
		OllamaBaseURL:   getEnv("OLLAMA_BASE_URL"),
		OllamaModel:     getEnv("OLLAMA_MODEL"),
	}
}

// normalizeConfigPaths expands and validates paths in the config. If createWorkDir is true, it creates the work directory if it doesn't exist.
func normalizeConfigPaths(cfg *Config, createWorkDir bool) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.AgentWorkDir != "" {
		workDir, err := normalizePath(cfg.AgentWorkDir)
		if err != nil {
			return err
		}
		cfg.AgentWorkDir = workDir

		if err := ensureWorkDir(workDir, createWorkDir); err != nil {
			return err
		}
	}

	return nil
}

// normalizePath expands ~ to home directory and cleans the path
func normalizePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	return filepath.Clean(path), nil
}

// ensureWorkDir checks if the given path exists and is a directory. If create is true, it creates the directory if it doesn't exist.
func ensureWorkDir(path string, create bool) error {
	if path == "" {
		return nil
	}

	if create {
		return os.MkdirAll(path, 0700)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("AGENT_WORKDIR must be a directory: %s", path)
	}

	return nil
}

// defaultWorkDir returns the default agent work directory path, which is ~/.vayuu/workspace
func defaultWorkDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".vayuu", "workspace"), nil
}
