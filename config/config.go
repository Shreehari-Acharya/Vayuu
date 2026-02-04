package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	TelegramToken string
	ApiKey        string
	ApiBaseURL    string
	Model         string
	AgentWorkDir  string
}

var (
	instance *Config
	once     sync.Once
)

// Load loads and validates configuration
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		// Load .env file (ignore error if not exists)
		_ = godotenv.Load()

		instance = &Config{
			TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
			ApiKey:        os.Getenv("API_KEY"),
			ApiBaseURL:    os.Getenv("API_BASE_URL"),
			Model:         os.Getenv("MODEL"),
			AgentWorkDir:  os.Getenv("AGENT_WORKDIR"),
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

	return nil
}
