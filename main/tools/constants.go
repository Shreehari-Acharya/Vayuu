package tools

import "time"

// Command execution limits
const (
	MaxCommandTimeout = 30 * time.Second
	MaxCommands       = 20
)

// File size limits
const (
	MaxReadFileSize  = 5 * 1024 * 1024  // 5MB
	MaxAppendSize    = 5 * 1024 * 1024  // 5MB
	MaxTotalFileSize = 50 * 1024 * 1024 // 50MB
	MaxCommandOutput = 10 * 1024 * 1024 // 10MB
)

// File operation limits
const (
	MaxFilesPerOperation = 50
)

// Protected files from deletion
var ProtectedPatterns = []string{
	".env",
	"config.json",
	".git/",
}
