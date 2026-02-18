package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/agent"
	"github.com/Shreehari-Acharya/vayuu/internal/prompts"
	"github.com/Shreehari-Acharya/vayuu/internal/telegram"
	"github.com/Shreehari-Acharya/vayuu/internal/tools"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if len(os.Args) > 1 && os.Args[1] == "setup" {
		if err := config.RunSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nRun 'vayuu setup' to configure the application.")
		os.Exit(1)
	}

	agentInstance, err := agent.CreateAgent(prompts.SystemPrompt, cfg)
	if err != nil {
		slog.Error("failed to create agent", "error", err)
		os.Exit(1)
	}

	toolEnv, err := tools.NewToolEnv(cfg.AgentWorkDir)
	if err != nil {
		slog.Error("failed to initialize tool environment", "error", err)
		os.Exit(1)
	}

	if err := tools.RegisterAll(toolEnv, agentInstance); err != nil {
		slog.Error("failed to register tools", "error", err)
		os.Exit(1)
	}

	bot, err := telegram.NewBot(cfg, agentInstance, toolEnv)
	if err != nil {
		slog.Error("failed to create telegram bot", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go bot.Start(ctx)
	slog.Info("bot is running — send a message on Telegram to interact")

	<-ctx.Done()
	slog.Info("shutdown signal received, stopping...")
	time.Sleep(500 * time.Millisecond)
	slog.Info("shutdown complete")
}
