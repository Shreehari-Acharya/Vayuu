package tools

import "github.com/Shreehari-Acharya/vayuu/main/agent"

func ExecuteCommandTool() agent.Tool {
	return agent.Tool{
		Name:        "execute_command",
		Description: "Execute bash command(s) on the local system",
		Parameters: map[string]any{
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
		Handler: ExecuteCommand,
	}
}

func ReadFileTool() agent.Tool {
	return agent.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file(s) at the given path",
		Parameters: map[string]any{
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

func AppendToFileTool() agent.Tool {
	return agent.Tool{
		Name:        "append_to_file",
		Description: "Append content to a file at the given path (will create the file if it does not exist)",
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
		Handler: AppendToFile,
	}
}

func DeleteFileTool() agent.Tool {
	return agent.Tool{
		Name:        "delete_file",
		Description: "Delete  file(s) at the given path",
		Parameters: map[string]any{
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
		Handler: DeleteFile,
	}
}


func GetAllTools() []agent.Tool { 
    return []agent.Tool{
        ExecuteCommandTool(),
        ReadFileTool(),
        WriteFileTool(),
		AppendToFileTool(),
		DeleteFileTool(),
    }
}

