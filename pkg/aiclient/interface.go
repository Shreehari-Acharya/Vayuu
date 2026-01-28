package aiclient

import "context"

// Message represents a single message in a conversation.
type Message struct {
	Role    string
	Content string
}

// ToolDefinition defines a tool that an AI can invoke.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCall represents a tool invocation made by the AI.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Response contains the AI's reply and any tool calls.
type Response struct {
	Content   string
	ToolCalls []ToolCall
}

// Client defines the interface for communicating with an AI service.
type Client interface {
	// Complete sends a message and returns the response.
	Complete(ctx context.Context, messages []Message) (string, error)

	// CompleteWithTools sends a message with available tools and returns response + tool calls.
	CompleteWithTools(ctx context.Context, systemPrompt string, tools []ToolDefinition) (Response, error)
}
