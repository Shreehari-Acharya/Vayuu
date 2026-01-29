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

func ReadFileTool() agent.Tool {
	return agent.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file at the given path",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"path"},
		},
		Handler: ReadFile,
	}
}

func WriteFileTool() agent.Tool {
	return agent.Tool{
		Name:        "write_file",
		Description: "Write content to a file at the given path",
		Parameters: map[string]any{
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
		Handler: WriteFile,
	}
}


func GetAllTools() []agent.Tool { 
    return []agent.Tool{
        ExecuteCommandTool(),
        ReadFileTool(),
        WriteFileTool(),
    }
}

