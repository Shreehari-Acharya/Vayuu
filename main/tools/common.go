package tools
// common functions used by many tools

import (
	"os"
	"path/filepath"
	"strings"
)

func expandPath(path string) string {
    if !strings.HasPrefix(path, "~") {
        return filepath.Clean(path)
    }
    
    home, err := os.UserHomeDir()
    if err != nil {
        return filepath.Clean(path) // Fallback to literal if home is missing
    }
    
    // Join home with everything after the ~
    return filepath.Join(home, path[1:])
}