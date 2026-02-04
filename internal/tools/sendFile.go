package tools

import (
	"fmt"
	"os"
)

// SendFile handles sending a file to the user
// Returns a special marker that triggers file sending
func SendFile(args map[string]any) string {
	pathStr, ok := args["path"].(string)
	if !ok {
		return "Error: path must be a string"
	}

	caption, _ := args["caption"].(string)

	// Validate path
	fullPath, err := ValidatePath(pathStr)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Validate file exists and is not a directory
	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Sprintf("Error: file not found: %v", err)
	}

	if IsFileDirectory(fullPath) {
		return "Error: path is a directory, not a file"
	}

	// Use the file sender function to send the file
	if fileSender == nil {
		return "Error: file sender not configured"
	}

	err = fileSender(fullPath, caption)
	if err != nil {
		return fmt.Sprintf("Error sending file: %v", err)
	}

	return fmt.Sprintf("File sent: %s", pathStr)
}
