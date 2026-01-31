package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/main/agent"
	"github.com/Shreehari-Acharya/vayuu/main/telegram"
	"github.com/Shreehari-Acharya/vayuu/main/prompts"
	"github.com/Shreehari-Acharya/vayuu/main/tools"
)


var agentInstance *agent.Agent
var bot *telegram.Bot


func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// create an agent
	agentInstance = agent.NewAgent(prompts.SystemPrompt, cfg)

	// assign tools to the agent
	tools.Initialize(cfg, agentInstance)

	// start telegram bot with the agent
	bot, err = telegram.NewBot(cfg, &ctx, agentInstance)
	if err != nil {
		panic(err)
	}

	bot.Start(ctx)
	
}
