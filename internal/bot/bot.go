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

// Bot represents the Telegram bot instance.
type Bot struct {
	bot     *gotgbot.Bot
	updater *ext.Updater
}

// New creates and initializes a new Telegram bot with the provided token and service.
func New(token string, ai aiclient.Client, conversation *memory.Store) (*Bot, error) {
	tgBot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		return nil, err
	}

	// Create service and message handler
	service := NewService(ai, conversation)
	handler := NewMessageHandler(service)

	// Setup dispatcher with command and message handlers
	dispatcher := ext.NewDispatcher(nil)
	dispatcher.AddHandler(handlers.NewCommand("start", handler.HandleStart))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, handler.HandleMessage))

	updater := ext.NewUpdater(dispatcher, nil)

	return &Bot{
		bot:     tgBot,
		updater: updater,
	}, nil
}

// Start begins polling for updates and handles graceful shutdown.
func (b *Bot) Start() error {
	log.Println("bot starting...")

	err := b.updater.StartPolling(b.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 30,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: 50 * time.Second,
			},
		},
	})
	if err != nil {
		return err
	}

	log.Println("bot is running. press ctrl+c to stop")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down bot...")
	b.updater.Stop()
	log.Println("bot stopped gracefully")

	return nil
}
