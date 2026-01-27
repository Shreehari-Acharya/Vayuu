package main

import (
	"log"
	"time"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/ai"
	"github.com/Shreehari-Acharya/vayuu/internal/bot"
)

func main() {
	cfg := config.Get()
    var aiService ai.AIService
    var err error
	// Initialize AI
	if cfg.UseGroq {
        aiService, err = ai.NewGroq(cfg.GroqKey)
        if err != nil {
            log.Fatal(err)
        }
    } else {
        aiService, err = ai.NewGemini(cfg.GeminiKey)
        if err != nil {
            log.Fatal(err)
        }
    }

	// Initialize Handlers
	h := &bot.Handler{AI: aiService}

	// Initialize Telegram
	b, err := gotgbot.NewBot(cfg.TelegramToken, nil)
    if err != nil {
        log.Panic(err)
    }
	dispatcher := ext.NewDispatcher(nil)

	dispatcher.AddHandler(handlers.NewCommand("start", h.Start))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, h.HandleMessage))

	updater := ext.NewUpdater(dispatcher, nil)
	log.Println("Bot is running...")
	updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 30,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 50,
			},
		},
	})
	updater.Idle()
}
