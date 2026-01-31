package tools

import (
    "os"

	"github.com/Shreehari-Acharya/vayuu/config"
    "github.com/Shreehari-Acharya/vayuu/main/agent"
)

// agentWorkDir is the working directory for all agent operations
// It's initialized from the config on package load
var (
    agentWorkDir string
    fileSender  func(filePath, caption string) error
    currentChatID int64
)

func Initialize(cfg *config.Config, agent *agent.Agent) {
	if cfg != nil {
		agentWorkDir = cfg.AgentWorkDir
	}

    // register all tools at initialization
    tools := getAllTools()
    for _, tool := range tools {
        // Assuming there's a global agent instance to register tools with
        agent.RegisterTool(tool)
    }

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
	fileSender = sender
}

func SetCurrentChatID(chatID int64) {
	currentChatID = chatID
}