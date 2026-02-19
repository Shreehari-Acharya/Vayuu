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

// sendMessage sends a text message to the current chat ID using the Telegram bot API, with MarkdownV1 parsing enabled.
func (tb *Bot) sendMessage(ctx context.Context, text string) error {
	_, err := tb.bot.SendMessage(ctx, &bot.SendMessageParams{
		ParseMode: models.ParseModeMarkdownV1,
		ChatID:    tb.currentChatID,
		Text:      text,
	})
	return err
}

// SendContent is a public method that allows sending various types of content (images, videos, documents) to the current chat. It detects the content type, validates it, and calls the appropriate method to send the content using the Telegram bot API.
func (tb *Bot) sendTypingAction(ctx context.Context) error {
	_, err := tb.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: tb.currentChatID,
		Action: models.ChatActionTyping,
	})
	return err
}

// SendContent is a public method that allows sending various types of content (images, videos, documents) to the current chat. It detects the content type, validates it, and calls the appropriate method to send the content using the Telegram bot API.
func (tb *Bot) SendContent(content, caption string) error {
	if tb.currentChatID == 0 {
		return fmt.Errorf("no active chat")
	}

	detectedType, err := DetectContentType(content)
	if err != nil {
		return fmt.Errorf("detect content type: %w", err)
	}

	contentType, err := ValidateContentType("", detectedType)
	if err != nil {
		slog.Warn("content type validation failed, using detected", "error", err)
		contentType = detectedType
	}

	slog.Debug("sending content", "type", contentType, "content_len", len(content))

	switch contentType {
	case ContentTypeImage:
		return tb.sendPhoto(content, caption)
	case ContentTypeVideo:
		return tb.sendVideo(content, caption)
	case ContentTypeDoc:
		return tb.sendDocument(content, caption)
	default:
		return tb.sendDocument(content, caption)
	}
}

// sendPhoto sends an image file to the current chat ID using the Telegram bot API, with an optional caption.
func (tb *Bot) sendPhoto(filePath, caption string) error {
	if err := validateFileForUpload(filePath); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	params := &bot.SendPhotoParams{
		ChatID: tb.currentChatID,
		Photo: &models.InputFileUpload{
			Filename: filepath.Base(filePath),
			Data:     f,
		},
	}
	if caption != "" {
		params.Caption = caption
	}

	if _, err := tb.bot.SendPhoto(context.Background(), params); err != nil {
		return fmt.Errorf("send photo: %w", err)
	}

	slog.Info("photo sent", "chat_id", tb.currentChatID, "path", filePath)
	return nil
}

// sendVideo sends a video file to the current chat ID using the Telegram bot API, with an optional caption.
func (tb *Bot) sendVideo(filePath, caption string) error {
	if err := validateFileForUpload(filePath); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	params := &bot.SendVideoParams{
		ChatID: tb.currentChatID,
		Video: &models.InputFileUpload{
			Filename: filepath.Base(filePath),
			Data:     f,
		},
	}
	if caption != "" {
		params.Caption = caption
	}

	if _, err := tb.bot.SendVideo(context.Background(), params); err != nil {
		return fmt.Errorf("send video: %w", err)
	}

	slog.Info("video sent", "chat_id", tb.currentChatID, "path", filePath)
	return nil
}

// sendDocument sends a document file to the current chat ID using the Telegram bot API, with an optional caption.
func (tb *Bot) sendDocument(filePath, caption string) error {
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

	if _, err := tb.bot.SendDocument(context.Background(), params); err != nil {
		return fmt.Errorf("send document: %w", err)
	}

	slog.Info("document sent", "chat_id", tb.currentChatID, "path", filePath)
	return nil
}
