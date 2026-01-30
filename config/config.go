package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	TelegramToken string
	ApiKey        string
	ApiBaseURL    string
	Provider      string
	MemorySize    int
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
			Provider:      os.Getenv("PROVIDER"),
			Model:         os.Getenv("MODEL"),
			MemorySize:    getEnvAsInt("MEMORY_SIZE", 20),
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

	if c.Provider == "" {
		return fmt.Errorf("PROVIDER is required")
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

	if c.MemorySize <= 0 {
		return fmt.Errorf("MEMORY_SIZE must be positive")
	}

	return nil
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
