package bot

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/Shreehari-Acharya/vayuu/internal/ai"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"log"
	"time"
)

type Handler struct {
	STM *memory.STM
	AI ai.AIService
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Hi! I'm Vayuu. Your AI assistant.", nil)
	return err
}

func (h *Handler) HandleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Printf("Received message from %s: %s", ctx.EffectiveUser.Username, ctx.EffectiveMessage.Text)
	userText := ctx.EffectiveMessage.Text

	h.STM.Add("user", userText)
	history := h.STM.GetHistory()
	b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)

	
	aiCtx, aiCancel := context.WithTimeout(context.Background(), 30*time.Second)
	response, err := h.AI.Ask(aiCtx, userText, history)
	aiCancel()

	h.STM.Add("assistant", response)

	if err != nil {
		log.Printf("AI Error: %v", err)
		response = "The AI is taking too long to think. Please try again."
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdown,
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: 15 * time.Second, 
		},
	}

	_, err = b.SendMessage(ctx.EffectiveChat.Id, response, opts)

	if err != nil {
		log.Printf("Markdown failed, sending plain text: %v", err)
		//fallback to plain text
		_, err = b.SendMessage(ctx.EffectiveChat.Id, response, &gotgbot.SendMessageOpts{
			RequestOpts: &gotgbot.RequestOpts{Timeout: 30 * time.Second},
		})
	}

	return err
}
