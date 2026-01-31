package tools

import (
	"fmt"
	"os"
	"strings"
)

// DeleteFile handles deleting files
// Supports both single file (string) and multiple files (array)
func DeleteFile(args map[string]any) string {
	// Handle single file (string)
	if pathStr, ok := args["path"].(string); ok {
		return deleteSingleFile(pathStr)
	}

	// Handle multiple files (array)
	if pathArray, ok := args["path"].([]any); ok {
		return deleteMultipleFiles(pathArray)
	}

	return "Error: path must be a string or array of strings"
}

func deleteMultipleFiles(pathArray []any) string {
	if len(pathArray) == 0 {
		return "Error: path array cannot be empty"
	}

	if len(pathArray) > MaxFilesPerOperation {
		return fmt.Sprintf("Error: too many files (max %d)", MaxFilesPerOperation)
	}

	var results []string
	successCount := 0
	failCount := 0

	for i, path := range pathArray {
		pathStr, ok := path.(string)
		if !ok {
			results = append(results, fmt.Sprintf("Error: path[%d] must be a string", i))
			failCount++
			continue
		}

		result := deleteSingleFile(pathStr)
		if strings.HasPrefix(result, "Error:") {
			results = append(results, fmt.Sprintf("%s: %s", pathStr, result))
			failCount++
		} else {
			results = append(results, result)
			successCount++
		}
	}

	summary := fmt.Sprintf("Deleted %d file(s), %d failed", successCount, failCount)
	if len(results) > 0 {
		return fmt.Sprintf("%s\n\n%s", summary, strings.Join(results, "\n"))
	}
	return summary
}

func deleteSingleFile(relativePath string) string {
	fullPath, err := ValidatePath(relativePath)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// Check if file exists
	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Error: file does not exist: %s", relativePath)
		}
		return fmt.Sprintf("Error: cannot access file: %v", err)
	}

	// Prevent deleting directories (use a separate tool for that)
	if IsFileDirectory(fullPath) {
		return fmt.Sprintf("Error: path is a directory, not a file: %s", relativePath)
	}

	// Optional: Prevent deleting important files
	if IsProtectedFile(relativePath) {
		return fmt.Sprintf("Error: cannot delete protected file: %s", relativePath)
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Sprintf("Error deleting file: %v", err)
	}

	return fmt.Sprintf("Successfully deleted: %s", relativePath)
}
