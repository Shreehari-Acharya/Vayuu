package tools

import (
	"fmt"
	"os/exec"
	"context"
	"time"
	"strings"
	"os"
)

// The Tool Handler function to execute shell commands
const (
	maxCommandTimeout = 30 * time.Second
	maxOutputSize     = 10 * 1024 * 1024 // 10MB
	maxCommands       = 20                
)

func ExecuteCommand(args map[string]any) string {
	// Handle single command (string)
	if cmdStr, ok := args["command"].(string); ok {
		return executeSingleCommand(cmdStr)
	}

	// Handle multiple commands (array)
	if cmdArray, ok := args["command"].([]any); ok {
		return executeMultipleCommands(cmdArray)
	}

	return "Error: command must be a string or array of strings"
}

func executeMultipleCommands(cmdArray []any) string {
	if len(cmdArray) == 0 {
		return "Error: command array cannot be empty"
	}

	if len(cmdArray) > maxCommands {
		return fmt.Sprintf("Error: too many commands (max %d)", maxCommands)
	}

	var results []string
	for i, cmd := range cmdArray {
		cmdStr, ok := cmd.(string)
		if !ok {
			return fmt.Sprintf("Error: command[%d] must be a string", i)
		}

		if strings.TrimSpace(cmdStr) == "" {
			return fmt.Sprintf("Error: command[%d] cannot be empty", i)
		}

		result := executeSingleCommand(cmdStr)
		
		// Check if command failed - optionally stop on first error
		if strings.HasPrefix(result, "Error:") {
			results = append(results, fmt.Sprintf("Command %d failed: %s\n%s", i+1, cmdStr, result))
			// Uncomment to stop on first error:
			// return strings.Join(results, "\n\n")
		} else {
			results = append(results, fmt.Sprintf("=== Command %d: %s ===\n%s", i+1, cmdStr, result))
		}
	}

	return strings.Join(results, "\n\n")
}

func executeSingleCommand(cmd string) string {
	if strings.TrimSpace(cmd) == "" {
		return "Error: command cannot be empty"
	}

	fmt.Printf("Executing command in %s: %s\n", agentWorkDir, cmd)

	ctx, cancel := context.WithTimeout(context.Background(), maxCommandTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, "bash", "-c", cmd)
	command.Dir = agentWorkDir

	if err := validateWorkDir(); err != nil {
		return fmt.Sprintf("Error: invalid work directory: %v", err)
	}

	output, err := command.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: command timed out after %v", maxCommandTimeout)
	}

	if len(output) > maxOutputSize {
		return fmt.Sprintf("Error: command output too large (%.2f MB, max 10 MB)\nFirst 1000 chars:\n%s",
			float64(len(output))/(1024*1024),
			string(output[:1000]))
	}

	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}

	return string(output)
}

func validateWorkDir() error {
	info, err := os.Stat(agentWorkDir)
	if err != nil {
		return fmt.Errorf("cannot access: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	return nil
}
