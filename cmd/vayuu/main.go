package main

import (
	"log"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/bot"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

// run initializes and starts the application.
func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Create AI client
	aiClient, err := aiclient.New(aiclient.Config{
		UseGroq:   cfg.UseGroq,
		GroqKey:   cfg.GroqKey,
		GeminiKey: cfg.GeminiKey,
	})
	if err != nil {
		return err
	}

	// Create conversation store
	conversationStore := memory.New(cfg.MemorySize)

	// Create and start bot
	bot, err := bot.New(cfg.TelegramToken, aiClient, conversationStore)
	if err != nil {
		return err
	}

	return bot.Start()
}
