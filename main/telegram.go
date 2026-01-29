package main

import (
	"context"
	"log"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func startTelegramBot(ctx *context.Context, cfg *config.Config) {

	opts := []bot.Option{
		bot.WithDefaultHandler(onMessage),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if nil != err {
		log.Fatal(err.Error())
	}

	b.Start(*ctx)
}

func onMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		
		// call agent with message text and chat id, and bot instance
		// (why bot instance? to send message containing docs or other stuff)
		// get response from agent
		agentResponse := "This is a placeholder response from the agent."

		// send response back to user
		err := sendMessage(update.Message.Chat.ID, agentResponse, b, ctx)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}

func sendMessage(chatID int64, text string, b *bot.Bot, ctx context.Context) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})

	return err
}