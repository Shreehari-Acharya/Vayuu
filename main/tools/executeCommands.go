package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteCommand handles executing shell commands
// Supports both single command (string) and multiple commands (array)

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

	if len(cmdArray) > MaxCommands {
		return fmt.Sprintf("Error: too many commands (max %d)", MaxCommands)
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

	// fmt.Printf("Executing command in %s: %s\n", agentWorkDir, cmd)

	ctx, cancel := context.WithTimeout(context.Background(), MaxCommandTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, "bash", "-c", cmd)
	command.Dir = agentWorkDir

	if err := ValidateWorkDir(); err != nil {
		return fmt.Sprintf("Error: invalid work directory: %v", err)
	}

	output, err := command.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: command timed out after %v", MaxCommandTimeout)
	}

	if len(output) > MaxCommandOutput {
		return fmt.Sprintf("Error: command output too large (%s, max %s)\nFirst 1000 chars:\n%s",
			FormatBytes(int64(len(output))), FormatBytes(MaxCommandOutput),
			string(output[:1000]))
	}

	if err != nil {
		return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
	}

	return string(output)
}
