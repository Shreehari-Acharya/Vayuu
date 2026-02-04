package tools

import (
	"fmt"
	"os"
	"strings"
)

// ReadFile handles reading file contents
// Supports both single file (string) and multiple files (array)
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
	fullPath, err := ValidatePath(relativePath)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if IsFileDirectory(fullPath) {
		return "Error: path is a directory, not a file"
	}

	fileSize, err := GetFileSize(fullPath)
	if err != nil {
		return fmt.Sprintf("Error accessing file: %v", err)
	}

	if fileSize > MaxReadFileSize {
		return fmt.Sprintf("Error: file too large (%s, max %s)",
			FormatBytes(fileSize), FormatBytes(MaxReadFileSize))
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	return string(data)
}
