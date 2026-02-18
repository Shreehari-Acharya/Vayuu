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

type Bot struct {
	bot           *bot.Bot
	agent         *agent.Agent
	cfg           *config.Config
	toolEnv       *tools.ToolEnv
	currentChatID int64
}

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

	toolEnv.SetFileSender(tb.sendFileToCurrentChat)
	slog.Info("telegram bot initialized")

	return tb, nil
}

func (tb *Bot) Start(ctx context.Context) {
	slog.Info("telegram bot started, listening for messages")
	tb.bot.Start(ctx)
}

func (tb *Bot) setCurrentChatID(chatID int64) {
	tb.currentChatID = chatID
	tb.toolEnv.SetCurrentChatID(chatID)
}
