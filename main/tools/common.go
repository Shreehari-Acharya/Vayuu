package tools

import (
	"github.com/Shreehari-Acharya/vayuu/config"
)

// agentWorkDir is the working directory for all agent operations
// It's initialized from the config on package load
var agentWorkDir string

func init() {
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		agentWorkDir = cfg.AgentWorkDir
	}
}

// GetAgentWorkDir returns the working directory for agent operations
func GetAgentWorkDir() string {
	return agentWorkDir
}
