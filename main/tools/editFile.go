package tools

import (
	"fmt"
	"os"
	"strings"
)

// EditFile handles editing files by replacing strings
// Takes an old string to find and a new string to replace it with
func EditFile(args map[string]any) string {
	path, okPath := args["path"].(string)
	oldString, okOld := args["old_string"].(string)
	newString, okNew := args["new_string"].(string)

	if !okPath || !okOld || !okNew {
		return "Error: 'path', 'old_string', and 'new_string' must be strings"
	}

	fullPath, err := ValidatePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Read the file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}

	fileContent := string(content)

	// Check if the old string exists
	if !strings.Contains(fileContent, oldString) {
		return "Error: old_string not found in file. Please check the exact string including whitespace and line breaks."
	}

	// Count occurrences
	occurrences := strings.Count(fileContent, oldString)
	if occurrences > 1 {
		return fmt.Sprintf("Error: old_string appears %d times in the file. Please provide a more specific string that appears only once.", occurrences)
	}

	// Replace the string
	newContent := strings.Replace(fileContent, oldString, newString, 1)

	// Write back to file
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err)
	}

	// Calculate change statistics
	oldLines := strings.Count(oldString, "\n") + 1
	newLines := strings.Count(newString, "\n") + 1
	sizeDiff := len(newContent) - len(fileContent)

	var sizeChange string
	if sizeDiff > 0 {
		sizeChange = fmt.Sprintf("+%s", FormatBytes(int64(sizeDiff)))
	} else if sizeDiff < 0 {
		sizeChange = fmt.Sprintf("-%s", FormatBytes(int64(-sizeDiff)))
	} else {
		sizeChange = "no size change"
	}

	return fmt.Sprintf("Successfully edited %s\nReplaced %d line(s) with %d line(s) (%s)",
		path, oldLines, newLines, sizeChange)
}
