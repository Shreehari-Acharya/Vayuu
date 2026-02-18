package tools

func buildToolDefs(env *ToolEnv) []toolDef {
	return []toolDef{
		{
			name:        "read_file",
			description: "Read the contents of one or more files at the given path(s)",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []string{"path"},
			},
			handler: env.readFile,
		},
		{
			name:        "write_file",
			description: "Write content to a file at the given path",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []string{"path", "content"},
			},
			handler: env.writeFile,
		},
		{
			name:        "execute_command",
			description: "Execute bash command(s) on the local system",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"required": []string{"command"},
			},
			handler: env.executeCommand,
		},
		{
			name:        "send_file",
			description: "Send a file to the user via Telegram",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    map[string]any{"type": "string", "description": "Path to the file to send"},
					"caption": map[string]any{"type": "string", "description": "Optional caption for the file"},
				},
				"required": []string{"path"},
			},
			handler: env.sendFile,
		},
		{
			name:        "edit_file",
			description: "Edit a file by replacing an exact string match with a new string. The old_string must appear exactly once.",
			parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":       map[string]any{"type": "string", "description": "Path to the file to edit"},
					"old_string": map[string]any{"type": "string", "description": "Exact string to find (must match once)"},
					"new_string": map[string]any{"type": "string", "description": "Replacement string"},
				},
				"required": []string{"path", "old_string", "new_string"},
			},
			handler: env.editFile,
		},
	}
}
