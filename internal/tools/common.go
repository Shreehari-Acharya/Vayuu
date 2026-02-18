package tools

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/Shreehari-Acharya/vayuu/internal/agent"
)

type FileSenderFunc func(filePath, caption string) error

type ToolEnv struct {
	WorkDir       string
	FileSender    FileSenderFunc
	CurrentChatID int64
	mu            sync.RWMutex
}

func NewToolEnv(workDir string) (*ToolEnv, error) {
	if workDir == "" {
		return nil, fmt.Errorf("work directory must not be empty")
	}
	info, err := os.Stat(workDir)
	if err != nil {
		return nil, fmt.Errorf("cannot access work directory %q: %w", workDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("work directory is not a directory: %s", workDir)
	}
	return &ToolEnv{WorkDir: workDir}, nil
}

func (e *ToolEnv) SetFileSender(sender FileSenderFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.FileSender = sender
}

func (e *ToolEnv) SetCurrentChatID(chatID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CurrentChatID = chatID
}

func (e *ToolEnv) getFileSender() FileSenderFunc {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.FileSender
}

func RegisterAll(env *ToolEnv, a *agent.Agent) error {
	for _, def := range buildToolDefs(env) {
		tool := agent.Tool{
			Name:        def.name,
			Description: def.description,
			Parameters:  def.parameters,
			Handler:     def.handler,
		}
		if err := a.RegisterTool(tool); err != nil {
			return fmt.Errorf("register tool %q: %w", def.name, err)
		}
		slog.Debug("registered tool", "name", def.name)
	}
	return nil
}
