package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

type Bot struct {
	bot     *gotgbot.Bot
	updater *ext.Updater
	handler *Handler
}

// Initialize sets up the Telegram bot with the provided token, AI service, and short-term memory.
func Initialize(token string, ai aiclient.AIService, stm *memory.STM) (*Bot, error) {
	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		return nil, err
	}

	// Store handler as pointer
	handler := &Handler{
		STM: stm,
		AI:  ai,
	}

	dispatcher := ext.NewDispatcher(nil)
	dispatcher.AddHandler(handlers.NewCommand("start", handler.Start))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, handler.HandleMessage))

	updater := ext.NewUpdater(dispatcher, nil)

	return &Bot{
		bot:     b,
		updater: updater,
		handler: handler,
	}, nil
}

// Start begins polling for updates and handles graceful shutdown on interrupt signals.
func (b *Bot) Start() error {
	log.Println("Bot is starting...")

	err := b.updater.StartPolling(b.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 30,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 50,
			},
		},
	})
	if err != nil {
		return err
	}

	log.Println("Bot is running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot...")
	b.updater.Stop()
	log.Println("Bot stopped gracefully")

	return nil
}
