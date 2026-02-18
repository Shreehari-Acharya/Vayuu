package tools

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func (e *ToolEnv) executeCommand(args map[string]any) string {
	if cmdStr, ok := args["command"].(string); ok {
		return e.runCommand(cmdStr)
	}

	if cmdArray, ok := args["command"].([]any); ok {
		return e.runMultipleCommands(cmdArray)
	}

	return "error: command must be a string or array of strings"
}

func (e *ToolEnv) runMultipleCommands(cmds []any) string {
	if len(cmds) == 0 {
		return "error: command array is empty"
	}
	if len(cmds) > maxCommands {
		return fmt.Sprintf("error: too many commands (max %d)", maxCommands)
	}

	var results []string
	for i, cmd := range cmds {
		cmdStr, ok := cmd.(string)
		if !ok {
			return fmt.Sprintf("error: command[%d] must be a string", i)
		}
		if strings.TrimSpace(cmdStr) == "" {
			return fmt.Sprintf("error: command[%d] is empty", i)
		}

		result := e.runCommand(cmdStr)
		if strings.HasPrefix(result, "error:") {
			results = append(results, fmt.Sprintf("command %d failed: %s\n%s", i+1, cmdStr, result))
		} else {
			results = append(results, fmt.Sprintf("=== command %d: %s ===\n%s", i+1, cmdStr, result))
		}
	}
	return strings.Join(results, "\n\n")
}

func (e *ToolEnv) runCommand(cmd string) string {
	if strings.TrimSpace(cmd) == "" {
		return "error: command is empty"
	}

	slog.Debug("executing command", "dir", e.WorkDir, "cmd", cmd)

	ctx, cancel := context.WithTimeout(context.Background(), maxCommandTimeout)
	defer cancel()

	proc := exec.CommandContext(ctx, "bash", "-c", cmd)
	proc.Dir = e.WorkDir

	output, err := proc.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("error: command timed out after %v", maxCommandTimeout)
	}

	if len(output) > maxCommandOutput {
		return fmt.Sprintf("error: output too large (%s, max %s)\nfirst 1000 chars:\n%s",
			formatBytes(int64(len(output))), formatBytes(int64(maxCommandOutput)),
			string(output[:1000]))
	}

	if err != nil {
		return fmt.Sprintf("error: %v\noutput: %s", err, string(output))
	}

	return string(output)
}
