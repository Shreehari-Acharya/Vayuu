package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const maxTelegramFileSize = 50 * 1024 * 1024

func (tb *Bot) sendMessage(ctx context.Context, text string) error {
	_, err := tb.bot.SendMessage(ctx, &bot.SendMessageParams{
		ParseMode: models.ParseModeMarkdownV1,
		ChatID:    tb.currentChatID,
		Text:      text,
	})
	return err
}

func (tb *Bot) sendTypingAction(ctx context.Context) error {
	_, err := tb.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: tb.currentChatID,
		Action: models.ChatActionTyping,
	})
	return err
}

func (tb *Bot) sendFileToCurrentChat(filePath, caption string) error {
	if tb.currentChatID == 0 {
		return fmt.Errorf("no active chat")
	}
	return tb.sendDocument(context.Background(), filePath, caption)
}

func (tb *Bot) sendDocument(ctx context.Context, filePath, caption string) error {
	if err := validateFileForUpload(filePath); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	params := &bot.SendDocumentParams{
		ChatID: tb.currentChatID,
		Document: &models.InputFileUpload{
			Filename: filepath.Base(filePath),
			Data:     f,
		},
	}
	if caption != "" {
		params.Caption = caption
	}

	if _, err := tb.bot.SendDocument(ctx, params); err != nil {
		return fmt.Errorf("send document: %w", err)
	}

	slog.Info("file sent", "chat_id", tb.currentChatID, "path", filePath)
	return nil
}

func validateFileForUpload(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}
	if info.Size() > maxTelegramFileSize {
		return fmt.Errorf("file too large (%.2f MB, max 50 MB)", float64(info.Size())/(1024*1024))
	}
	return nil
}
