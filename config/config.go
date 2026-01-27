package config

import (
	"log"
	"os"
	"sync"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	GeminiKey     string
	GroqKey       string
	UseGroq	  	  bool
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		_ = godotenv.Load()
		instance = &Config{
			TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
			GeminiKey:     os.Getenv("GEMINI_API_KEY"),
			GroqKey:       os.Getenv("GROQ_API_KEY"),
			UseGroq:       os.Getenv("USE_GROQ") == "true",
		}
		if instance.TelegramToken == "" || instance.GeminiKey == "" {
			log.Fatal("Missing critical environment variables")
		}
	})
	return instance
}