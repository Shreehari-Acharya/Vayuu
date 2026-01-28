package aiclient

import (
	"context"
)

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string
	Content string
}

// AIService defines the interface for an AI service.
type AIService interface {
	Ask(ctx context.Context, history []ChatMessage) (string, error)
}
