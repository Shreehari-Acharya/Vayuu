package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryManager struct {
	embedder *Embedder
	store    *VectorStore
	config   *Config

	mu          sync.RWMutex
	memoryCount int
}

func NewMemoryManager(cfg *Config) (*MemoryManager, error) {
	embedder := NewEmbedder(cfg)

	store, err := NewVectorStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("create vector store: %w", err)
	}

	return &MemoryManager{
		embedder: embedder,
		store:    store,
		config:   cfg,
	}, nil
}

func (m *MemoryManager) AddMemory(ctx context.Context, content string, memType MemoryType, metadata map[string]string) error {
	start := time.Now()

	vector, err := m.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	id := uuid.New().String()
	createdAt := time.Now().Format(time.RFC3339)

	payload := map[string]any{
		"content":    content,
		"type":       string(memType),
		"created_at": createdAt,
	}

	for k, v := range metadata {
		payload[k] = v
	}

	if err := m.store.Upsert(ctx, id, vector, payload); err != nil {
		return fmt.Errorf("store memory: %w", err)
	}

	m.mu.Lock()
	m.memoryCount++
	m.mu.Unlock()

	slog.Debug("memory added", "id", id, "type", memType, "duration", time.Since(start))
	return nil
}

func (m *MemoryManager) SearchMemory(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	start := time.Now()

	vector, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generate embedding: %w", err)
	}

	results, err := m.store.Search(ctx, vector, limit, nil)
	if err != nil {
		return nil, fmt.Errorf("search memory: %w", err)
	}

	slog.Debug("memory searched", "query", query, "results", len(results), "duration", time.Since(start))
	return results, nil
}

func (m *MemoryManager) GetContext(ctx context.Context, query string, maxTokens int) (string, error) {
	results, err := m.SearchMemory(ctx, query, 10)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	var contextParts []string
	currentLen := 0

	for _, r := range results {
		mem := r.Memory
		memText := fmt.Sprintf("[%s] %s", mem.Type, mem.Content)

		if currentLen+len(memText) > maxTokens*4 {
			break
		}

		contextParts = append(contextParts, memText)
		currentLen += len(memText)
	}

	if len(contextParts) == 0 {
		return "", nil
	}

	return "Relevant memories:\n" + strings.Join(contextParts, "\n\n"), nil
}

func (m *MemoryManager) AddFact(ctx context.Context, fact string, metadata map[string]string) error {
	return m.AddMemory(ctx, fact, MemoryTypeFact, metadata)
}

func (m *MemoryManager) AddPreference(ctx context.Context, preference string, metadata map[string]string) error {
	return m.AddMemory(ctx, preference, MemoryTypePreference, metadata)
}

func (m *MemoryManager) AddKnowledge(ctx context.Context, knowledge string, metadata map[string]string) error {
	return m.AddMemory(ctx, knowledge, MemoryTypeKnowledge, metadata)
}

func (m *MemoryManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryCount
}

func (m *MemoryManager) Close() error {
	return m.store.Close()
}
