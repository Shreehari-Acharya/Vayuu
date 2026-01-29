package main

// create telegram instance

// onMessage handler to process incoming messages
// it should be then sent to a agent which uses llm and memory to respond

// start the telegram bot

// Send any text message to the bot after the bot has been started

import (
	"context"
	"os"
	"os/signal"
	
	"github.com/Shreehari-Acharya/vayuu/config"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	startTelegramBot(&ctx, cfg)
}
