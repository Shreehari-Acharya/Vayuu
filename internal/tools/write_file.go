package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeFile is a tool function that writes content to a file specified by its relative path. It validates the file path, creates necessary directories, and writes the content to the file. The function returns a success message with the number of bytes written or any errors encountered during the process.
func (e *ToolEnv) writeFile(args map[string]any) string {
	path, okPath := args["path"].(string)
	content, okContent := args["content"].(string)

	if !okPath || !okContent {
		return "error: 'path' and 'content' must be strings"
	}

	fullPath, err := e.validatePath(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Sprintf("error creating directory: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("error writing file: %v", err)
	}

	return fmt.Sprintf("wrote %s to %s", formatBytes(int64(len(content))), path)
}
