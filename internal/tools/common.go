package tools

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Shreehari-Acharya/vayuu/internal/agent"
)

// ToolEnv provides a shared environment for tools, allowing them to access common resources such as the work directory and a file sender function for sending content back to the user.
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

// SetFileSender sets the FileSender function in the ToolEnv, allowing tools to send files back to the user through the Telegram bot. This method is thread-safe, ensuring that concurrent access to the FileSender is properly synchronized.
func (e *ToolEnv) SetFileSender(sender FileSenderFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.FileSender = sender
}

// SetCurrentChatID updates the current chat ID in both the Bot struct and the tool environment, ensuring that tools have access to the correct chat context when sending messages or files. This method is thread-safe, allowing concurrent updates to the current
func (e *ToolEnv) SetCurrentChatID(chatID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CurrentChatID = chatID
}

// getFileSender retrieves the current FileSender function from the ToolEnv in a thread-safe manner, allowing tools to access the function for sending files back to the user through the Telegram bot.
func (e *ToolEnv) getFileSender() FileSenderFunc {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.FileSender
}

// RegisterAll registers all available tools in the provided ToolEnv with the given Agent instance. It iterates through the tool definitions, creates Tool instances, and registers them with the agent, logging the registration process and returning any errors encountered during registration.
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
