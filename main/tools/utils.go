package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePath ensures a file path is within the agent work directory (prevents path traversal)
func ValidatePath(relativePath string) (string, error) {
	if strings.TrimSpace(relativePath) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	fullPath := filepath.Join(agentWorkDir, relativePath)

	// Security: Prevent path traversal
	cleanPath := filepath.Clean(fullPath)
	cleanWorkDir := filepath.Clean(agentWorkDir)
	if !strings.HasPrefix(cleanPath, cleanWorkDir) {
		return "", fmt.Errorf("path traversal not allowed")
	}

	return fullPath, nil
}

// ValidateWorkDir ensures the agent work directory exists and is valid
func ValidateWorkDir() error {
	info, err := os.Stat(agentWorkDir)
	if err != nil {
		return fmt.Errorf("cannot access: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	return nil
}

// IsFileDirectory checks if a path points to a directory
func IsFileDirectory(fullPath string) bool {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize returns the size of a file
func GetFileSize(fullPath string) (int64, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// IsProtectedFile checks if a file is protected from operations like deletion
func IsProtectedFile(path string) bool {
	for _, pattern := range ProtectedPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}
