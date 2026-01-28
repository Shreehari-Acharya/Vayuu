package bot

import (
	"context"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// MessageHandler handles Telegram messages and commands.
type MessageHandler struct {
	service *Service
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(service *Service) *MessageHandler {
	return &MessageHandler{service: service}
}

// HandleStart responds to the /start command.
func (h *MessageHandler) HandleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	welcomeMsg := "Hi! I'm Vayuu, your AI assistant. How can I help you today?"
	_, err := ctx.EffectiveMessage.Reply(b, welcomeMsg, nil)
	if err != nil {
		log.Printf("failed to send start message: %v", err)
	}
	return err
}

// HandleMessage processes user messages and sends AI responses.
func (h *MessageHandler) HandleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	userText := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveChat.Id
	username := ctx.EffectiveUser.Username

	log.Printf("message from @%s: %s", username, userText)

	// Send typing indicator
	if err := h.sendTypingAction(b, chatID); err != nil {
		log.Printf("failed to send typing action: %v", err)
	}

	// Process message and get response
	response, err := h.service.ProcessMessage(context.Background(), userText)
	if err != nil {
		log.Printf("failed to process message: %v", err)
		response = "I'm having trouble processing your request. Please try again."
	}

	// Send response to user
	return h.sendResponse(b, chatID, response)
}

// sendTypingAction sends a typing action to indicate the bot is processing.
func (h *MessageHandler) sendTypingAction(b *gotgbot.Bot, chatID int64) error {
	_, err := b.SendChatAction(chatID, "typing", nil)
	return err
}

// sendResponse sends the AI response to the user with markdown formatting fallback.
func (h *MessageHandler) sendResponse(b *gotgbot.Bot, chatID int64, text string) error {
	opts := &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdown,
	}

	// Try with markdown first
	_, err := b.SendMessage(chatID, text, opts)
	if err != nil {
		log.Printf("markdown parsing failed, sending as plain text: %v", err)

		// Fallback to plain text
		_, err = b.SendMessage(chatID, text, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
