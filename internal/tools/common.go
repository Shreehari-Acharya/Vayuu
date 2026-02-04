package tools

import (
	"fmt"
	"os"
	"sync"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/agent"
)

// agentWorkDir is the working directory for all agent operations
// It's initialized from the config on package load
var (
	agentWorkDir  string
	fileSender    func(filePath, caption string) error
	currentChatID int64
	mu            sync.RWMutex // Protect concurrent access
)

func Initialize(cfg *config.Config, agent *agent.Agent) error {
	if cfg != nil {
		agentWorkDir = cfg.AgentWorkDir
	}

	// Validate work directory before registering tools
	if err := ValidateWorkDir(); err != nil {
		return fmt.Errorf("invalid work directory: %w", err)
	}

	// register all tools at initialization
	tools := getAllTools()
	for _, tool := range tools {
		if err := agent.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}

	return nil
}

func handleTildeInPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return homeDir + path[1:]
		}
	}
	return "failed to get user home directory"
}

// GetAgentWorkDir returns the working directory for agent operations
func GetAgentWorkDir() string {
	return agentWorkDir
}

func SetFileSender(sender func(filePath, caption string) error) {
	mu.Lock()
	defer mu.Unlock()
	fileSender = sender
}

func SetCurrentChatID(chatID int64) {
	mu.Lock()
	defer mu.Unlock()
	currentChatID = chatID
}

// ValidateWorkDir validates that the work directory exists and is accessible
func ValidateWorkDir() error {
	if agentWorkDir == "" {
		return fmt.Errorf("work directory not set")
	}

	info, err := os.Stat(agentWorkDir)
	if err != nil {
		return fmt.Errorf("cannot access work directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("work directory is not a directory: %s", agentWorkDir)
	}

	return nil
}
