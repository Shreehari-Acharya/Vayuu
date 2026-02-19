package tools

import (
	"fmt"
	"os"
)

// sendFile is a tool function that sends a file to a chat using the configured FileSender. It validates the file path, checks if the file exists and is not a directory, and then uses the FileSender to send the file with an optional caption. The function returns a success message or any errors encountered during the process.
func (e *ToolEnv) sendFile(args map[string]any) string {
	pathStr, ok := args["path"].(string)
	if !ok {
		return "error: path must be a string"
	}

	caption, _ := args["caption"].(string)

	fullPath, err := e.validatePath(pathStr)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Sprintf("error: file not found: %v", err)
	}

	if isDirectory(fullPath) {
		return "error: path is a directory, not a file"
	}

	sender := e.getFileSender()
	if sender == nil {
		return "error: file sender not configured"
	}

	if err := sender(fullPath, caption); err != nil {
		return fmt.Sprintf("error sending content: %v", err)
	}

	return fmt.Sprintf("content sent: %s", pathStr)
}
