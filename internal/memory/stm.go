package memory

import (
	"sync"

	"github.com/Shreehari-Acharya/vayuu/pkg/aiclient"
)

// Store manages conversation history with a sliding window approach.
// It is safe for concurrent use.
type Store struct {
	mu      sync.RWMutex
	history []aiclient.Message
	maxSize int
}

// New creates a new conversation store with the specified maximum size.
// If maxSize is <= 0, it defaults to 20.
func New(maxSize int) *Store {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &Store{
		history: make([]aiclient.Message, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a message to the conversation history.
// If the history exceeds maxSize, the oldest pair (user + assistant) is removed.
func (s *Store) Add(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, aiclient.Message{
		Role:    role,
		Content: content,
	})

	// Remove oldest pair if exceeding max size
	if len(s.history) > s.maxSize {
		s.history = s.history[2:]
	}
}

// GetHistory returns a copy of the current conversation history.
func (s *Store) GetHistory() []aiclient.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]aiclient.Message, len(s.history))
	copy(history, s.history)
	return history
}

// Clear removes all messages from the conversation history.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = s.history[:0]
}

// Len returns the current number of messages in the conversation history.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.history)
}
