package tools

import "time"

const (
	maxCommandTimeout = 30 * time.Second
	maxCommands       = 20
	maxReadFileSize   = 5 * 1024 * 1024
	maxTotalFileSize  = 50 * 1024 * 1024
	maxCommandOutput  = 10 * 1024 * 1024
)
