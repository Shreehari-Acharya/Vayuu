package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile handles writing content to a file
// Creates the file or overwrites it if it exists
func WriteFile(args map[string]any) string {
	path, okPath := args["path"].(string)
	content, okContent := args["content"].(string)

	if !okPath || !okContent {
		return "Error: 'path' and 'content' must be strings"
	}

	fullPath, err := ValidatePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := EnsureDirectoryExists(dir); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	// Write the file (0644 = readable by all, writable by owner)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err)
	}

	return fmt.Sprintf("Successfully wrote %s to %s", FormatBytes(int64(len(content))), path)
}
