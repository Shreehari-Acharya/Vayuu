package tools

import (
	"fmt"
	"os"
	"strings"
)

// AppendToFile handles appending content to a file
// Creates the file if it does not exist
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
	fullPath, err := ValidatePath(relativePath)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Check content size
	if len(content) > MaxAppendSize {
		return fmt.Sprintf("Error: content too large (%s, max %s)",
			FormatBytes(int64(len(content))), FormatBytes(MaxAppendSize))
	}

	// Check if file exists and get its size
	var currentSize int64
	if info, err := os.Stat(fullPath); err == nil {
		// File exists
		if IsFileDirectory(fullPath) {
			return fmt.Sprintf("Error: path is a directory, not a file: %s", relativePath)
		}
		currentSize = info.Size()

		// Check total size after append
		newSize := currentSize + int64(len(content))
		if newSize > MaxTotalFileSize {
			return fmt.Sprintf("Error: file would be too large after append (%s, max %s)",
				FormatBytes(newSize), FormatBytes(MaxTotalFileSize))
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

	totalSize := currentSize + int64(bytesWritten)
	return fmt.Sprintf("Successfully %s file: %s (%d bytes written, total size: %s)",
		action, relativePath, bytesWritten, FormatBytes(totalSize))
}
