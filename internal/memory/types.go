package memory

import (
	"time"

	"github.com/openai/openai-go/v3"
)

// Configuration constants for memory storage
const (
	DefaultMemoryMaxSize = 10 * 1024 * 1024 // 10MB - max file size before rotation
	MemoryDirName        = "memory"         // Subdirectory name for memory files
	DayFileLayout        = "2006-01-02"     // Date format for daily files
	ClockLayout          = "15:04:05"       // Time format for timestamps
)

// MemoryType categorizes memories for filtering and retrieval.
// Different types can be queried separately for specific context.
type MemoryType string

// Memory type constants for categorization
const (
	MemoryTypeFact         MemoryType = "fact"         // Factual information about user
	MemoryTypePreference   MemoryType = "preference"   // User preferences (food, hobbies, etc.)
	MemoryTypeKnowledge    MemoryType = "knowledge"    // General knowledge learned about user
	MemoryTypeConversation MemoryType = "conversation" // Full conversation logs
)

// Memory represents a single memory entry stored in the vector database.
// Each memory has content, a type category, and metadata for filtering.
type Memory struct {
	ID        string            // Unique identifier (UUID)
	Content   string            // The actual memory content
	Type      MemoryType        // Category of memory (fact, preference, etc.)
	Metadata  map[string]string // Additional key-value metadata
	CreatedAt string            // RFC3339 timestamp when created
}

// SearchResult contains a memory and its similarity score from vector search.
// Higher score means more similar to the query.
type SearchResult struct {
	Memory Memory  // The matched memory
	Score  float64 // Similarity score (0-1, higher is better)
}

// Config holds configuration for the memory system.
// Default values are provided via DefaultConfig().
type Config struct {
	OllamaBaseURL  string // URL of Ollama server for embeddings
	OllamaModel    string // Model name for generating embeddings
	QdrantURL      string // URL of Qdrant vector database
	VectorDim      int    // Dimension of embedding vectors
	CollectionName string // Name of the Qdrant collection
}

// MemoryWriter is the interface for persisting conversation history.
// Implementations can store to files, databases, etc.
type MemoryWriter interface {
	Write(messages []openai.ChatCompletionMessageParamUnion) error
}

// FileMemoryWriter stores conversation history in daily JSONL files.
// Each line is a JSON object with timestamp, role, and content.
// Files are rotated when they exceed MaxSize.
type FileMemoryWriter struct {
	Dir     string           // Directory path for memory files
	MaxSize int64            // Max file size before rotation (bytes)
	Clock   func() time.Time // Clock for testing (defaults to time.Now)
}

// MemoryEntry represents a single message in conversation history.
// Stored in JSONL format for easy appending and parsing.
type MemoryEntry struct {
	Timestamp string `json:"timestamp"` // Time in HH:MM:SS format
	Role      string `json:"role"`      // "user" or "assistant"
	Content   string `json:"content"`   // Message content
}

// DefaultConfig returns a Config with sensible defaults for local development.
// Override values by modifying the returned Config before passing to New* functions.
func DefaultConfig() *Config {
	return &Config{
		OllamaBaseURL:  "http://localhost:11434",
		OllamaModel:    "nomic-embed-text",
		QdrantURL:      "localhost:6334",
		VectorDim:      768, // Dimension for nomic-embed-text
		CollectionName: "vayuu_memory",
	}
}
