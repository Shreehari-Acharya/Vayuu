package memory

import "sync"

type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

type STM struct {
	mu      sync.Mutex
	History []Message
	MaxSize int
}

func NewSTM(size int) *STM {
	return &STM{
		History: make([]Message, 0),
		MaxSize: size,
	}
}

func (s *STM) Add(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.History = append(s.History, Message{Role: role, Content: content})
	
	// Sliding window: If history is too long, remove the oldest pair
	if len(s.History) > s.MaxSize {
		s.History = s.History[2:] // Removes the oldest User+AI turn
	}
}

func (s *STM) GetHistory() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message(nil), s.History...) // Return a copy to avoid race conditions
}

func (s *STM) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = []Message{}
}