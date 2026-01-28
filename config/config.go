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
	GeminiKey     string
	GroqKey       string
	UseGroq       bool
	MemorySize    int
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
			GeminiKey:     os.Getenv("GEMINI_API_KEY"),
			GroqKey:       os.Getenv("GROQ_API_KEY"),
			UseGroq:       os.Getenv("USE_GROQ") == "true",
			MemorySize:    getEnvAsInt("MEMORY_SIZE", 20),
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

	if c.UseGroq {
		if c.GroqKey == "" {
			return fmt.Errorf("GROQ_API_KEY is required when USE_GROQ=true")
		}
	} else {
		if c.GeminiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required when USE_GROQ=false")
		}
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
