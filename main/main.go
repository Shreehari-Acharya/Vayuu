package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/main/agent"
	"github.com/Shreehari-Acharya/vayuu/main/prompts"
	"github.com/Shreehari-Acharya/vayuu/main/telegram"
	"github.com/Shreehari-Acharya/vayuu/main/tools"
)

var agentInstance *agent.Agent
var bot *telegram.Bot

func main() {
	// Check for setup command
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		if err := config.RunSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nRun 'vayuu setup' to configure the application.")
		os.Exit(1)
	}

	// create an agent
	agentInstance = agent.NewAgent(prompts.SystemPrompt, cfg)

	// assign tools to the agent
	if err := tools.Initialize(cfg, agentInstance); err != nil {
		panic(fmt.Errorf("failed to initialize tools: %w", err))
	}

	// start telegram bot with the agent
	bot, err = telegram.NewBot(cfg, &ctx, agentInstance)
	if err != nil {
		panic(err)
	}

	// Start bot in goroutine
	go bot.Start(ctx)

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	fmt.Println("\nShutting down gracefully...")
	time.Sleep(1 * time.Second)
	fmt.Println("Shutdown complete")

}
