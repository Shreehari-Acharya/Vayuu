package bot

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

const (
	defaultAITimeout     = 30 * time.Second
	defaultSendTimeout   = 15 * time.Second
)

// Handler handles Telegram bot commands and messages
type Handler struct {
	STM *memory.STM
	AI  aiclient.AIService
}

// Start handles the /start command
func (h *Handler) Start(b *gotgbot.Bot, ctx *ext.Context) error {
	welcomeMsg := "Hi! I'm Vayuu, your AI assistant. How can I help you today?"
	_, err := ctx.EffectiveMessage.Reply(b, welcomeMsg, nil)
	if err != nil {
		log.Printf("Failed to send start message: %v", err)
	}
	return err
}

// HandleMessage processes user messages and generates AI responses
// first id adds user message to STM, then gets AI response, adds it to STM and sends it back to user
func (h *Handler) HandleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	userText := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveChat.Id

	log.Printf("[%s] Message: %s", ctx.EffectiveUser.Username, userText)

	// Add user message to memory
	h.STM.Add("user", userText)

	// Send typing action
	if err := h.sendTypingAction(b, chatID); err != nil {
		log.Printf("Failed to send typing action: %v", err)
	}

	// Get AI response
	response, err := h.getAIResponse()
	if err != nil {
		log.Printf("AI error: %v", err)
		response = "I'm having trouble processing your request. Please try again."
	}

	// Add assistant response to memory
	h.STM.Add("assistant", response)

	// Send response to user
	return h.sendResponse(b, chatID, response)
}

func (h *Handler) sendTypingAction(b *gotgbot.Bot, chatID int64) error {
	_, err := b.SendChatAction(chatID, "typing", nil)
	return err
}

// getAIResponse retrieves the AI response based on the current conversation history.
// the history should contain the user query as the last message
func (h *Handler) getAIResponse() (string, error) {
	// Get history - already a copy with pre-allocated capacity
	history := h.STM.GetHistory()

	// Call AI with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultAITimeout)
	defer cancel()

	response, err := h.AI.Ask(ctx, history)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}

	return response, nil
}

// sendResponse sends the AI response to the user, trying Markdown first and falling back to plain text if needed.
// This ensures better formatting while handling potential parsing issues.
func (h *Handler) sendResponse(b *gotgbot.Bot, chatID int64, text string) error {
	opts := &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdown,
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: defaultSendTimeout,
		},
	}

	// Try sending with Markdown formatting
	_, err := b.SendMessage(chatID, text, opts)
	if err != nil {
		log.Printf("Markdown parsing failed, sending as plain text: %v", err)

		// Fallback to plain text
		plainOpts := &gotgbot.SendMessageOpts{
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: defaultSendTimeout,
			},
		}
		_, err = b.SendMessage(chatID, text, plainOpts)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}
