package agent

import (
	"github.com/openai/openai-go/v3"
)

// Logf is a lightweight logger signature compatible with fmt.Printf.
type Logf func(format string, args ...any) (int, error)

// MemoryWriter persists conversation history for later use.
// Implementations should be safe for repeated calls and handle their own storage concerns.
type MemoryWriter interface {
	Write(messages []openai.ChatCompletionMessageParamUnion) error
}

// Agent represents an AI agent with tool calling capabilities.
type Agent struct {
	client       *openai.Client
	model        string
	tools        map[string]Tool
	toolsCache   []openai.ChatCompletionToolUnionParam
	toolsDirty   bool
	systemPrompt string
	workDir      string
	memoryWriter MemoryWriter
	logf         Logf
}

// MemoryEntry represents a single message written to memory storage.
type MemoryEntry struct {
	Timestamp string `json:"timestamp"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

// ToolFunc is the handler signature for agent tools.
type ToolFunc func(args map[string]any) string

// Tool defines a tool that can be invoked by the agent via function calling.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema
	Handler     ToolFunc
}
