package config

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/term"
)

// Config holds all application configuration
type Config struct {
	TelegramToken string
	ApiKey        string
	ApiBaseURL    string
	Model         string
	AgentWorkDir  string
	AllowedUsername string
}

var (
	instance *Config
	once     sync.Once
)

// Load loads and validates configuration
// Tries encrypted config first, then falls back to environment variables
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		// Try to load from encrypted config if it exists
		if hasEncryptedConfig() {
			password := os.Getenv("VAYUU_PASSWORD")

			// If not set via env, try to get from system keyring
			if password == "" {
				password = TryGetKeystorePassword()
			}

			// If still no password, prompt user
			if password == "" {
				fmt.Print("Enter config password: ")
				var passBytes []byte
				passBytes, err = term.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					err = fmt.Errorf("failed to read password: %w", err)
					return
				}
				password = string(passBytes)
			}

			instance, err = LoadEncryptedConfig(password)
			return
		}

		// Fall back to .env file or environment variables
		_ = godotenv.Load()

		instance = &Config{
			TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
			ApiKey:        os.Getenv("API_KEY"),
			ApiBaseURL:    os.Getenv("API_BASE_URL"),
			Model:         os.Getenv("MODEL"),
			AgentWorkDir:  os.Getenv("AGENT_WORKDIR"),
			AllowedUsername: os.Getenv("ALLOWED_USERNAME"),
		}

		err = instance.validate()
	})

	if err != nil {
		return nil, err
	}
	return instance, nil
}

// Get returns the singleton config instance (deprecated: use Load instead)
func Get() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

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

	// Validate work directory exists and is accessible
	info, err := os.Stat(c.AgentWorkDir)
	if err != nil {
		return fmt.Errorf("AGENT_WORKDIR: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("AGENT_WORKDIR must be a directory: %s", c.AgentWorkDir)
	}

	if c.AllowedUsername == "" {
		return fmt.Errorf("ALLOWED_USERNAME is required")
	}

	return nil
}

// hasEncryptedConfig checks if encrypted config file exists
func hasEncryptedConfig() bool {
	configPath := getConfigPath()
	_, err := os.Stat(configPath)
	return err == nil
}
