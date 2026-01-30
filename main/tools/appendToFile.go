package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxAppendSize = 5 * 1024 * 1024 // 5MB max append size
	maxFileSize   = 50 * 1024 * 1024 // 50MB max total file size after append
)

func AppendToFile(args map[string]any) string {
	pathStr, ok := args["path"].(string)
	if !ok {
		return "Error: path must be a string"
	}

	content, ok := args["content"].(string)
	if !ok {
		return "Error: content must be a string"
	}

	addNewline, _ := args["newline"].(bool)
	if addNewline && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return appendToFile(pathStr, content)
}

func appendToFile(relativePath, content string) string {
	if strings.TrimSpace(relativePath) == "" {
		return "Error: path cannot be empty"
	}

	// Build full path
	fullPath := filepath.Join(agentWorkDir, relativePath)

	// Security: Prevent path traversal
	cleanPath := filepath.Clean(fullPath)
	cleanWorkDir := filepath.Clean(agentWorkDir)
	if !strings.HasPrefix(cleanPath, cleanWorkDir) {
		return "Error: path traversal not allowed"
	}

	// Check content size
	if len(content) > maxAppendSize {
		return fmt.Sprintf("Error: content too large (%.2f MB, max 5 MB)",
			float64(len(content))/(1024*1024))
	}

	// Check if file exists and get its size
	var currentSize int64
	if info, err := os.Stat(fullPath); err == nil {
		// File exists
		if info.IsDir() {
			return fmt.Sprintf("Error: path is a directory, not a file: %s", relativePath)
		}
		currentSize = info.Size()

		// Check total size after append
		newSize := currentSize + int64(len(content))
		if newSize > maxFileSize {
			return fmt.Sprintf("Error: file would be too large after append (%.2f MB, max 50 MB)",
				float64(newSize)/(1024*1024))
		}
	} else if !os.IsNotExist(err) {
		// Error other than file not existing
		return fmt.Sprintf("Error: cannot access file: %v", err)
	}
	// If file doesn't exist, it will be created

	// Open file for appending (create if doesn't exist)
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Sprintf("Error opening file: %v", err)
	}
	defer file.Close()

	// Write content
	bytesWritten, err := file.WriteString(content)
	if err != nil {
		return fmt.Sprintf("Error writing to file: %v", err)
	}

	fmt.Printf("Appended %d bytes to: %s\n", bytesWritten, fullPath)
	
	action := "appended to"
	if currentSize == 0 {
		action = "created and wrote to"
	}
	
	return fmt.Sprintf("Successfully %s file: %s (%d bytes written, total size: %.2f KB)",
		action, relativePath, bytesWritten, float64(currentSize+int64(bytesWritten))/1024)
}