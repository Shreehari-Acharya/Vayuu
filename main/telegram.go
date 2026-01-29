package main

import (
	"context"
	"log"

	"github.com/Shreehari-Acharya/vayuu/main/agent"
	"github.com/Shreehari-Acharya/vayuu/main/tools"
	"github.com/Shreehari-Acharya/vayuu/main/prompts"
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type App struct {
	agent *agent.Agent
}

func startTelegramBotWithAgent(ctx *context.Context, cfg *config.Config) {
	app := &App{
		agent: agent.NewAgent(prompts.GetSystemPrompt(), cfg),
	}

	// Register all tools using a loop
    for _, tool := range tools.GetAllTools() {
        app.agent.RegisterTool(tool)
    }

	opts := []bot.Option{
		bot.WithDefaultHandler(app.handler),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if nil != err {
		log.Fatal(err.Error())
	}

	b.Start(*ctx)
}

func (a* App) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	agentResponse, err  := a.agent.RunAgent(ctx, update.Message.Text)

	if err != nil {
		log.Printf("Error running agent: %v", err)
		return
	}

	err = sendMessage(update.Message.Chat.ID, agentResponse, b, ctx)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}	
}

func sendMessage(chatID int64, text string, b *bot.Bot, ctx context.Context) error {

	text = bot.EscapeMarkdownUnescaped(text)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ParseMode: models.ParseModeMarkdown,
		ChatID: chatID,
		Text:   text,
	})

	return err
}