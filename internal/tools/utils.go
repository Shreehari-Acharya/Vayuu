package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

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
