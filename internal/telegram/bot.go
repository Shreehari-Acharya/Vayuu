package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/agent"
	"github.com/Shreehari-Acharya/vayuu/internal/tools"
	"github.com/go-telegram/bot"
)

// Bot encapsulates the Telegram bot functionality, 
// integrating with the agent and tool environment to handle incoming messages and execute tools as needed.
func NewBot(cfg *config.Config, agentInstance *agent.Agent, toolEnv *tools.ToolEnv) (*Bot, error) {
	tb := &Bot{
		agent:   agentInstance,
		cfg:     cfg,
		toolEnv: toolEnv,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(tb.handleMessage),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	tb.bot = b

	toolEnv.SetFileSender(tb.SendContent)
	slog.Info("telegram bot initialized")

	return tb, nil
}

// Start begins the bot's message processing loop, allowing it to receive and respond to messages from users.
func (tb *Bot) Start(ctx context.Context) {
	slog.Info("telegram bot started, listening for messages")
	tb.bot.Start(ctx)
}

// setCurrentChatID updates the current chat ID in both the Bot struct and the tool environment, ensuring that tools have access to the correct chat context when sending messages or files.
func (tb *Bot) setCurrentChatID(chatID int64) {
	tb.currentChatID = chatID
	tb.toolEnv.SetCurrentChatID(chatID)
}
