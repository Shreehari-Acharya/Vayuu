package telegram

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// sendMessage sends a text message to a chat
func (tb *Bot) SendMessage(text string) error {
	ctx := *tb.ctx
	_, err := tb.bot.SendMessage(ctx, &bot.SendMessageParams{
		ParseMode: models.ParseModeMarkdownV1,
		ChatID:    tb.currentChatID,
		Text:      text,
	})
	return err
}

// sendTypingAction shows the typing indicator
func (tb *Bot) sendTypingAction() error {
	ctx := *tb.ctx
	_, err := tb.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: tb.currentChatID,
		Action: models.ChatActionTyping,
	})
	return err
}

// SendFileToCurrentChat sends a file to the current active chat
// This method is exposed to the tools package
func (tb *Bot) SendFileToCurrentChat(filePath, caption string) error {
	if tb.currentChatID == 0 {
		return fmt.Errorf("no active chat")
	}

	ctx := *tb.ctx

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	return tb.sendFile(filePath, caption)
}

// sendFile sends a file to a specific chat
func (tb *Bot) sendFile(filePath string, caption string) error {
	// Validate file
	if err := tb.validateFile(filePath); err != nil {
		return err
	}

	// Open the file
	fileData, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fileData.Close()

	filename := filepath.Base(filePath)

	params := &bot.SendDocumentParams{
		ChatID: tb.currentChatID,
		Document: &models.InputFileUpload{
			Filename: filename,
			Data:     fileData,
		},
	}

	if caption != "" {
		params.Caption = caption
	}

	// Use background context to avoid deadlock with parent context
	ctx := *tb.ctx
	_, err = tb.bot.SendDocument(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}

	log.Printf("Sent file to chat %d: %s", tb.currentChatID, filePath)
	return nil
}

// sendPhoto sends a photo to a chat
// func (tb *Bot) sendPhoto(chatID int64, photoPath string, caption string) error {
// 	fileData, err := os.Open(photoPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to open photo: %w", err)
// 	}
// 	defer fileData.Close()

// 	params := &bot.SendPhotoParams{
// 		ChatID: chatID,
// 		Photo: &models.InputFileUpload{
// 			Filename: filepath.Base(photoPath),
// 			Data:     fileData,
// 		},
// 	}

// 	if caption != "" {
// 		params.Caption = caption
// 	}

// 	_, err = tb.bot.SendPhoto(*tb.ctx, params)
// 	if err != nil {
// 		return fmt.Errorf("failed to send photo: %w", err)
// 	}

// 	return nil
// }

// sendMessageWithKeyboard sends a message with inline keyboard
// func (tb *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard [][]models.InlineKeyboardButton) error {
// 	_, err := tb.bot.SendMessage(*tb.ctx, &bot.SendMessageParams{
// 		ChatID: chatID,
// 		Text:   text,
// 		ReplyMarkup: &models.InlineKeyboardMarkup{
// 			InlineKeyboard: keyboard,
// 		},
// 	})
// 	return err
// }

// validateFile checks if a file is valid for sending
func (tb *Bot) validateFile(filePath string) error {

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	// Telegram bot file size limit is 50MB
	const maxFileSize = 50 * 1024 * 1024
	if info.Size() > maxFileSize {
		return fmt.Errorf("file too large (%.2f MB, max 50 MB)",
			float64(info.Size())/(1024*1024))
	}

	return nil
}
