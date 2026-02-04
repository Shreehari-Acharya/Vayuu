package tools

import "github.com/Shreehari-Acharya/vayuu/internal/agent"

// toolMetadata defines the structure for tool definitions
type toolMetadata struct {
	name        string
	description string
	parameters  map[string]any
	handler     func(map[string]any) string
}

// toolDefinitions holds all tool configurations
var toolDefinitions = []toolMetadata{
	{
		name:        "read_file",
		description: "Read the contents of a file(s) at the given path",
		parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			"required": []string{"path"},
		},
		handler: ReadFile,
	},
	{
		name:        "write_file",
		description: "Write content to a file at the given path",
		parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type": "string",
				},
				"content": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"path", "content"},
		},
		handler: WriteFile,
	},
	{
		name:        "execute_command",
		description: "Execute bash command(s) on the local system",
		parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			"required": []string{"command"},
		},
		handler: ExecuteCommand,
	},
	{
		name:        "send_file",
		description: "Send a file to the user",
		parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the file to send",
				},
				"caption": map[string]any{
					"type":        "string",
					"description": "Optional caption for the file",
				},
			},
			"required": []string{"path"},
		},
		handler: SendFile,
	},
	{
		name:        "edit_file",
		description: "Edit a file by replacing a specific string with a new string. The old_string must match exactly (including whitespace and line breaks) and appear only once in the file.",
		parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the file to edit",
				},
				"old_string": map[string]any{
					"type":        "string",
					"description": "The exact string to find and replace (must match exactly including whitespace)",
				},
				"new_string": map[string]any{
					"type":        "string",
					"description": "The new string to replace the old_string with",
				},
			},
			"required": []string{"path", "old_string", "new_string"},
		},
		handler: EditFile,
	},
}

// toAgentTool converts toolMetadata to agent.Tool
func (tm toolMetadata) toAgentTool() agent.Tool {
	return agent.Tool{
		Name:        tm.name,
		Description: tm.description,
		Parameters:  tm.parameters,
		Handler:     tm.handler,
	}
}

// GetAllTools returns all available tools as agent.Tool instances
func getAllTools() []agent.Tool {
	tools := make([]agent.Tool, len(toolDefinitions))
	for i, def := range toolDefinitions {
		tools[i] = def.toAgentTool()
	}
	return tools
}
