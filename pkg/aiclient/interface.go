package aiclient

import (
	"context"
)

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string
	Content string
}

// ToolDefinition represents a tool that the AI can call
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCall represents a tool call made by the AI
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// AIResponse contains the AI's response and any tool calls
type AIResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// AIService defines the interface for an AI service.
type AIService interface {
	// Ask sends a message and returns the response
	Ask(ctx context.Context, history []ChatMessage) (string, error)

	// AskWithTools sends a message with available tools and returns response + tool calls
	AskWithTools(ctx context.Context, context string, tools []ToolDefinition) (AIResponse, error)
}
