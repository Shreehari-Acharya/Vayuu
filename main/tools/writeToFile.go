package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// The Tool Handler function to write contents to a file
func WriteFile(args map[string]any) string {
	path, okPath := args["path"].(string)
	content, okContent := args["content"].(string)

	if !okPath || !okContent {
		return "Error: 'path' and 'content' must be strings"
	}

	path = expandPath(path)
	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)

	// Ensure the directory exists (0755 = standard directory permissions)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	// 3. Write the file (0644 = readable by all, writable by owner)
	// This will create the file or overwrite it if it exists.
	err := os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("Error writing file: %v", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), cleanPath)
}