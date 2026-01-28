package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

const (
	defaultCompletionTimeout = 30 * time.Second
)

// Service handles message processing with AI and conversation management.
type Service struct {
	ai           aiclient.Client
	conversation *memory.Store
}

// NewService creates a new bot service with the provided AI client and conversation store.
func NewService(ai aiclient.Client, conversation *memory.Store) *Service {
	return &Service{
		ai:           ai,
		conversation: conversation,
	}
}

// ProcessMessage processes a user message and returns the AI response.
// It adds the user message to history, gets the AI response, and updates history.
func (s *Service) ProcessMessage(ctx context.Context, userMessage string) (string, error) {
	// Add user message to conversation history
	s.conversation.Add("user", userMessage)

	// Get AI response
	response, err := s.getResponse(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get ai response: %w", err)
	}

	// Add assistant response to conversation history
	s.conversation.Add("assistant", response)

	return response, nil
}

// getResponse retrieves the AI response based on current conversation history.
func (s *Service) getResponse(ctx context.Context) (string, error) {
	// Use provided context or create one with timeout
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultCompletionTimeout)
		defer cancel()
	}

	history := s.conversation.GetHistory()

	response, err := s.ai.Complete(ctx, history)
	if err != nil {
		log.Printf("ai error: %v", err)
		return "", err
	}

	return response, nil
}

// ClearHistory clears the conversation history.
func (s *Service) ClearHistory() {
	s.conversation.Clear()
}
