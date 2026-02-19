package memory

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

func DefaultConfig() *Config {
	return &Config{
		OllamaBaseURL:  "http://localhost:11434",
		OllamaModel:    "nomic-embed-text",
		QdrantURL:      "localhost:6334",
		VectorDim:      768,
		CollectionName: "vayuu_memory",
	}
}
