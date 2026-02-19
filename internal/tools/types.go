package tools

import (
	"sync"
)

type FileSenderFunc func(content, caption string) error

type ToolEnv struct {
	WorkDir       string
	FileSender    FileSenderFunc
	CurrentChatID int64
	mu            sync.RWMutex
}

type toolDef struct {
	name        string
	description string
	parameters  map[string]any
	handler     func(map[string]any) string
}