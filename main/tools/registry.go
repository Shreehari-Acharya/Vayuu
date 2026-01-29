package tools

import "github.com/Shreehari-Acharya/vayuu/main/agent"

func ExecuteCommandTool() agent.Tool {
	return agent.Tool{
		Name:        "execute_command",
		Description: "Execute a shell (bash) command and return the output",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"command"},
		},
		Handler: ExecuteCommand,
	}
}