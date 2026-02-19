package tools

import (
	"fmt"
	"os"
	"strings"
)

// readFile is a tool function that reads the content of a file or multiple files specified by their paths. It validates the file paths, checks for file size limits, and returns the content of the file(s) or any errors encountered during the process. The function supports both single string paths and arrays of string paths, providing formatted output for multiple files.
func (e *ToolEnv) readFile(args map[string]any) string {
	if pathStr, ok := args["path"].(string); ok {
		return e.readSingleFile(pathStr)
	}

	if pathSlice, ok := args["path"].([]any); ok {
		var results []string
		for i, p := range pathSlice {
			pathStr, ok := p.(string)
			if !ok {
				return fmt.Sprintf("error: path[%d] must be a string", i)
			}
			results = append(results, fmt.Sprintf("=== %s ===\n%s", pathStr, e.readSingleFile(pathStr)))
		}
		return strings.Join(results, "\n\n")
	}

	return "error: path must be a string or array of strings"
}

// readSingleFile is a helper function that reads the content of a single file specified by its relative path. It validates the file path, checks if it's a directory, verifies the file size against the defined limit, and returns the file content or any errors encountered during the process.
func (e *ToolEnv) readSingleFile(relativePath string) string {
	fullPath, err := e.validatePath(relativePath)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	if isDirectory(fullPath) {
		return "error: path is a directory, not a file"
	}

	size, err := fileSize(fullPath)
	if err != nil {
		return fmt.Sprintf("error accessing file: %v", err)
	}
	if size > maxReadFileSize {
		return fmt.Sprintf("error: file too large (%s, max %s)", formatBytes(size), formatBytes(maxReadFileSize))
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}
	return string(data)
}
