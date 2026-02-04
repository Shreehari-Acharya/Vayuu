package telegram

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleMessage processes incoming messages
func (tb *Bot) handleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	
	if tb.cfg.AllowedUsername != "" && update.Message.From.Username != tb.cfg.AllowedUsername {
		log.Printf("Unauthorized user: %s", update.Message.From.Username)
		return
	}

	// Update current chat ID for tools to use
	tb.setCurrentChatID(update.Message.Chat.ID)

	// Send typing indicator
	if err := tb.sendTypingAction(); err != nil {
		log.Printf("Typing indicator failed: %v", err)
	}

	// Run agent
	agentResponse, err := tb.agent.RunAgent(ctx, update.Message.Text)
	if err != nil {
		log.Printf("Agent error: %v", err)
		tb.SendMessage("Sorry, I encountered an error processing your request.")
		return
	}

	// Send response
	if err := tb.SendMessage(agentResponse); err != nil {
		log.Printf("Send error: %v", err)
	}
}

// // handleCallback can be added for inline buttons
// func (tb *Bot) handleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	// Handle callback queries from inline keyboards
// 	if update.CallbackQuery == nil {
// 		return
// 	}

// 	// Implementation for callback handling
// 	log.Printf("Received callback: %s", update.CallbackQuery.Data)
// }

// // handleDocument can be added for file uploads
// func (tb *Bot) handleDocument(ctx context.Context, b *bot.Bot, update *models.Update) {
// 	if update.Message == nil || update.Message.Document == nil {
// 		return
// 	}

// 	// Handle document uploads
// 	log.Printf("Received document: %s", update.Message.Document.FileName)
// }