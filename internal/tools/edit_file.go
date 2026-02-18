package tools

import (
	"fmt"
	"os"
	"strings"
)

func (e *ToolEnv) editFile(args map[string]any) string {
	path, okPath := args["path"].(string)
	oldStr, okOld := args["old_string"].(string)
	newStr, okNew := args["new_string"].(string)

	if !okPath || !okOld || !okNew {
		return "error: 'path', 'old_string', and 'new_string' must be strings"
	}

	fullPath, err := e.validatePath(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}

	original := string(content)

	if !strings.Contains(original, oldStr) {
		return "error: old_string not found in file — check exact whitespace and line breaks"
	}

	if n := strings.Count(original, oldStr); n > 1 {
		return fmt.Sprintf("error: old_string appears %d times — provide a more specific match", n)
	}

	updated := strings.Replace(original, oldStr, newStr, 1)

	if err := os.WriteFile(fullPath, []byte(updated), 0644); err != nil {
		return fmt.Sprintf("error writing file: %v", err)
	}

	oldLines := strings.Count(oldStr, "\n") + 1
	newLines := strings.Count(newStr, "\n") + 1
	diff := len(updated) - len(original)

	var delta string
	switch {
	case diff > 0:
		delta = fmt.Sprintf("+%s", formatBytes(int64(diff)))
	case diff < 0:
		delta = fmt.Sprintf("-%s", formatBytes(int64(-diff)))
	default:
		delta = "no size change"
	}

	return fmt.Sprintf("edited %s: replaced %d line(s) with %d line(s) (%s)", path, oldLines, newLines, delta)
}
