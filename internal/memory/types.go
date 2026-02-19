package memory

import (
	"time"
	
	"github.com/openai/openai-go/v3"
)

type MemoryType string

const (
	MemoryTypeFact         MemoryType = "fact"
	MemoryTypePreference   MemoryType = "preference"
	MemoryTypeKnowledge    MemoryType = "knowledge"
	MemoryTypeConversation MemoryType = "conversation"
)

type Memory struct {
	ID        string
	Content   string
	Type      MemoryType
	Metadata  map[string]string
	CreatedAt string
}

type SearchResult struct {
	Memory Memory
	Score  float64
}

type Config struct {
	OllamaBaseURL  string
	OllamaModel    string
	QdrantURL      string
	VectorDim      int
	CollectionName string
}


type MemoryWriter interface {
	Write(messages []openai.ChatCompletionMessageParamUnion) error
}

// FileMemoryWriter appends conversation history to daily JSONL files.
type FileMemoryWriter struct {
	Dir     string
	MaxSize int64
	Clock   func() time.Time
}

type MemoryEntry struct {
	Timestamp string `json:"timestamp"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

func DefaultConfig() *Config {
	return &Config{
		OllamaBaseURL:  "http://localhost:11434",
		OllamaModel:    "nomic-embed-text",
		QdrantURL:      "localhost:6334",
		VectorDim:      768,
		CollectionName: "vayuu_memory",
	}
}
