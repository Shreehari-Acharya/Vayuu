package telegram

import (
	"context"
	"fmt"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/main/agent"
	"github.com/Shreehari-Acharya/vayuu/main/tools"
	"github.com/go-telegram/bot"
)

type Bot struct {
	bot           *bot.Bot
	agent         *agent.Agent
	currentChatID int64
	cfg 		 *config.Config
	ctx           *context.Context
}

// NewBot creates and initializes a new Telegram bot
func NewBot(cfg *config.Config, ctx *context.Context, agent *agent.Agent) (*Bot, error) {

	tb := &Bot{
		agent: agent,
		ctx:   ctx,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(tb.handleMessage),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	tb.bot = b

	tb.cfg = cfg
	// Inject file sender into tools
	tools.SetFileSender(tb.SendFileToCurrentChat)

	return tb, nil
}

// Start begins listening for messages
func (tb *Bot) Start(ctx context.Context) {
	tb.ctx = &ctx
	tb.bot.Start(ctx)
}

// GetCurrentChatID returns the current active chat ID
func (tb *Bot) GetCurrentChatID() int64 {
	return tb.currentChatID
}

// SetCurrentChatID updates the current chat ID
func (tb *Bot) setCurrentChatID(chatID int64) {
	tb.currentChatID = chatID
	tools.SetCurrentChatID(chatID)
}