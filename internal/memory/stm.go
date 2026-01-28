package memory

import (
	"sync"
	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

// STM (Short-Term Memory) manages conversation history
type STM struct {
	mu      sync.RWMutex
	history []aiclient.ChatMessage
	maxSize int
}

// NewSTM creates a new short-term memory with the specified size
func NewSTM(maxSize int) *STM {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &STM{
		history: make([]aiclient.ChatMessage, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a message to the history
func (s *STM) Add(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, aiclient.ChatMessage{
		Role:    role,
		Content: content,
	})

	// Sliding window: remove oldest messages if exceeding max size
	if len(s.history) > s.maxSize {
		// Remove oldest pair (user + assistant)
		s.history = s.history[2:]
	}
}

// GetHistory returns a copy of the conversation history
func (s *STM) GetHistory() []aiclient.ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]aiclient.ChatMessage, len(s.history))
	copy(history, s.history)
	return history
}

// Clear removes all messages from history
func (s *STM) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = s.history[:0]
}

// Len returns the current number of messages in history
func (s *STM) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.history)
}
