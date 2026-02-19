package agent

import (
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/openai/openai-go/v3"
)

type Agent struct {
	client       *openai.Client
	model        string
	tools        map[string]Tool
	toolsCache   []openai.ChatCompletionToolUnionParam
	toolsDirty   bool
	systemPrompt string
	workDir      string
	memoryWriter memory.MemoryWriter
	memoryMgr    *memory.MemoryManager
}

type ToolFunc func(args map[string]any) string

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	Handler     ToolFunc
}
