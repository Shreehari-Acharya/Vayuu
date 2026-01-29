package main

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

	startTelegramBotWithAgent(&ctx, cfg)
}
