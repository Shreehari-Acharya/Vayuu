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
			AgentWorkDir: os.Getenv("AGENT_WORKDIR"),
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

	return nil
}

