package tools
// common functions used by many tools

import (
    "github.com/Shreehari-Acharya/vayuu/config"
)

var agentWorkDir string

func init() {
    cfg, err := config.Load()
    if err == nil && cfg != nil {
        agentWorkDir = cfg.AgentWorkDir
    }
}
