package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleMessage is the main handler for incoming Telegram messages. It checks for allowed usernames, updates the current chat ID, sends a typing action, and processes the message using the agent, sending back the response to the user.
func (tb *Bot) handleMessage(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	username := update.Message.From.Username
	if tb.cfg.AllowedUsername != "" && username != tb.cfg.AllowedUsername {
		slog.Warn("rejected message from unauthorized user", "username", username)
		return
	}

	tb.setCurrentChatID(update.Message.Chat.ID)

	if err := tb.sendTypingAction(ctx); err != nil {
		slog.Debug("typing indicator failed", "error", err)
	}

	slog.Info("processing message", "user", username, "chat_id", update.Message.Chat.ID)

	response, err := tb.agent.RunAgent(ctx, update.Message.Text)
	if err != nil {
		slog.Error("agent failed", "error", err)
		_ = tb.sendMessage(ctx, "Sorry, I encountered an error processing your request.")
		return
	}

	if err := tb.sendMessage(ctx, response); err != nil {
		slog.Error("failed to send response", "error", err)
	}
}
