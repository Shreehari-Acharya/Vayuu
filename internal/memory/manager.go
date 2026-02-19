package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/google/uuid"
)

type MemoryManager struct {
	embedder  *Embedder
	store     *VectorStore
	database  *Database
	extractor *FactExtractor
	config    *Config

	mu          sync.RWMutex
	memoryCount int
}

func NewMemoryManager(cfg *Config) (*MemoryManager, error) {
	embedder := NewEmbedder(cfg)

	store, err := NewVectorStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("create vector store: %w", err)
	}

	mgr := &MemoryManager{
		embedder: embedder,
		store:    store,
		config:   cfg,
	}

	slog.Info("memory manager initialized (vector only)")
	return mgr, nil
}

func NewMemoryManagerWithDB(workDir string, cfg *config.Config) (*MemoryManager, error) {
	ollamaURL := cfg.OllamaBaseURL
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollamaModel := cfg.OllamaModel
	if ollamaModel == "" {
		ollamaModel = "nomic-embed-text"
	}

	memConfig := &Config{
		OllamaBaseURL:  ollamaURL,
		OllamaModel:    ollamaModel,
		CollectionName: "vayuu_memory",
		VectorDim:      768,
	}

	embedder := NewEmbedder(memConfig)

	store, err := NewVectorStore(memConfig)
	if err != nil {
		return nil, fmt.Errorf("create vector store: %w", err)
	}

	db, err := NewDatabase(workDir)
	if err != nil {
		slog.Warn("failed to initialize database", "error", err)
	}

	extractor := NewFactExtractor(cfg)

	mgr := &MemoryManager{
		embedder:  embedder,
		store:     store,
		database:  db,
		extractor: extractor,
		config:    memConfig,
	}

	slog.Info("memory manager initialized with database")
	return mgr, nil
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

	if len(results) == 0 && m.database == nil {
		return "", nil
	}

	var contextParts []string

	if m.database != nil {
		userSummary := m.database.GetUserSummary()
		if userSummary != "" {
			contextParts = append(contextParts, "User Profile:\n"+userSummary)
		}
	}

	for _, r := range results {
		mem := r.Memory
		memText := fmt.Sprintf("[%s] %s", mem.Type, mem.Content)

		if len(contextParts) > 0 && totalLen(contextParts)+len(memText) > maxTokens*4 {
			break
		}

		contextParts = append(contextParts, memText)
	}

	if len(contextParts) == 0 {
		return "", nil
	}

	return "Relevant memories:\n" + strings.Join(contextParts, "\n\n"), nil
}

func totalLen(parts []string) int {
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	return total
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

func (m *MemoryManager) ProcessConversation(ctx context.Context, userInput, assistantResponse string) error {
	if m.extractor == nil || m.database == nil {
		return nil
	}

	conversation := fmt.Sprintf("User: %s\nAssistant: %s", userInput, assistantResponse)

	facts, err := m.extractor.ExtractFacts(ctx, conversation)
	if err != nil {
		slog.Warn("failed to extract facts", "error", err)
		return nil
	}

	for _, fact := range facts {
		switch fact.Type {
		case "fact":
			if err := m.AddFact(ctx, fact.Value, map[string]string{"key": fact.Key}); err != nil {
				slog.Warn("failed to store fact", "error", err)
			}
			if fact.Key != "" && fact.Value != "" {
				m.database.SetProfile(fact.Key, fact.Value)
			}

		case "preference":
			if err := m.AddPreference(ctx, fact.Value, map[string]string{"key": fact.Key, "category": fact.Category}); err != nil {
				slog.Warn("failed to store preference", "error", err)
			}
			if fact.Key != "" && fact.Value != "" {
				m.database.SetPreference(fact.Key, fact.Value, fact.Category)
			}

		case "topic":
			m.database.IncrementTopic(fact.Value)
			if err := m.AddKnowledge(ctx, "Topic: "+fact.Value, map[string]string{"topic": fact.Value}); err != nil {
				slog.Warn("failed to store topic", "error", err)
			}
		}
	}

	return nil
}

func (m *MemoryManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryCount
}

func (m *MemoryManager) Close() error {
	if m.store != nil {
		m.store.Close()
	}
	if m.database != nil {
		m.database.Close()
	}
	return nil
}
