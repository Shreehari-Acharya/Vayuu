package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validatePath is a helper function that validates and resolves a relative file path against the ToolEnv's working directory. It checks for empty paths, handles paths starting with "~/", and ensures that the resolved path does not allow for path traversal outside of the working directory. The function returns the cleaned full path or an error if the validation fails.
func (e *ToolEnv) validatePath(relativePath string) (string, error) {
	if strings.TrimSpace(relativePath) == "" {
		return "", fmt.Errorf("path must not be empty")
	}

	if strings.HasPrefix(relativePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Clean(filepath.Join(home, relativePath[2:])), nil
	}

	fullPath := filepath.Clean(filepath.Join(e.WorkDir, relativePath))
	cleanWorkDir := filepath.Clean(e.WorkDir)
	if !strings.HasPrefix(fullPath, cleanWorkDir) {
		return "", fmt.Errorf("path traversal not allowed: %s", relativePath)
	}
	return fullPath, nil
}

// isDirectory checks if the given path is a directory. It returns true if the path exists and is a directory, and false otherwise.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}


// fileSize returns the size of the file at the given path. It returns the file size in bytes or an error if the file cannot be accessed.
func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// formatBytes is a helper function that formats a byte size into a human-readable string with appropriate units (B, KB, MB, GB). It takes the size in bytes and returns a formatted string with two decimal places for larger units.
func formatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
