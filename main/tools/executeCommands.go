package tools

import (
	"fmt"
	"os/exec"
)

// The Tool Handler function to execute shell commands
func ExecuteCommand(args map[string]any) (string) {
	cmd, ok := args["command"].(string)
	if !ok {
		return "Error: command must be a string"
	}
	
	osCmd := cmd + " 2>&1" // Capture stderr

	out, err := exec.Command("bash", "-c", osCmd).Output()
	if err != nil {
		return fmt.Sprintf("Error executing command: %s", err.Error())
	}
	return string(out)
}

