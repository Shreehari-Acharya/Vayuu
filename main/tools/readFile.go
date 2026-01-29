package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// The Tool Handler function to read file contents
func ReadFile(args map[string]any) string {
    path, ok := args["path"].(string)
    if !ok {
        return "Error: path must be a string"
    }

    path = expandPath(path)
    // Clean the path to prevent directory traversal
    cleanPath := filepath.Clean(path)

    // Check file stats before reading
    info, err := os.Stat(cleanPath)
    if err != nil {
        return fmt.Sprintf("Error: file not found: %s", err.Error())
    }

    // 3. Limit size (e.g., max 5MB) to prevent memory crashes
    if info.Size() > 5*1024*1024 {
        return "Error: file is too large (max 5MB)"
    }

    data, err := os.ReadFile(cleanPath)
    if err != nil {
        return fmt.Sprintf("Error reading file: %v", err)
    }

    return string(data)
}