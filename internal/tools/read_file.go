package tools

import (
	"fmt"
	"os"
	"strings"
)

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
