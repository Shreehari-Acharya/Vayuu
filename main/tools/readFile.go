package tools

import (
	"fmt"
	"os"
	"path/filepath"
    "strings"
)

// The Tool Handler function to read file contents
func ReadFile(args map[string]any) string {
	// Single file
	if pathStr, ok := args["path"].(string); ok {
		return readSingleFile(pathStr)
	}

	// Multiple files
	if pathSlice, ok := args["path"].([]any); ok {
		var results []string
		for i, p := range pathSlice {
			pathStr, ok := p.(string)
			if !ok {
				return fmt.Sprintf("Error: path[%d] must be a string", i)
			}
			result := readSingleFile(pathStr)
			results = append(results, fmt.Sprintf("=== %s ===\n%s", pathStr, result))
		}
		return strings.Join(results, "\n\n")
	}

	return "Error: path must be a string or array of strings"
}

func readSingleFile(relativePath string) string {
	fullPath := filepath.Join(agentWorkDir, relativePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Sprintf("Error accessing file: %v", err)
	}

	if info.IsDir() {
		return "Error: path is a directory, not a file"
	}

	const maxFileSize = 5 * 1024 * 1024
	if info.Size() > maxFileSize {
		return fmt.Sprintf("Error: file too large (%.2f MB, max 5 MB)", 
			float64(info.Size())/(1024*1024))
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	return string(data)
}