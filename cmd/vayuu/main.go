package main

import (
	"log"
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/bot"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

func main() {
	if err := vayuu(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func vayuu() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize AI service
	aiService, err := aiclient.NewAIService(cfg.UseGroq, cfg.GroqKey, cfg.GeminiKey)
	if err != nil {
		return err
	}

	// Initialize memory
	stm := memory.NewSTM(cfg.MemorySize)

	// Initialize bot 
	b, err := bot.Initialize(cfg.TelegramToken, aiService, stm)
	if err != nil {
		return err
	}

	// Start bot with graceful shutdown
	return b.Start()
}
