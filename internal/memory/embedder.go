package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ollama/ollama/api"
)

// Embedder generates text embeddings using Ollama's embedding API.
// It converts text into vector representations for semantic search.
type Embedder struct {
	baseURL string       // Ollama server URL (e.g., http://localhost:11434)
	model   string       // Model name for embeddings (e.g., nomic-embed-text)
	client  *http.Client // HTTP client for API requests
}

// NewEmbedder creates a new Embedder with the given configuration.
func NewEmbedder(cfg *Config) *Embedder {
	return &Embedder{
		baseURL: cfg.OllamaBaseURL,
		model:   cfg.OllamaModel,
		client:  &http.Client{},
	}
}

// Embed generates a vector embedding for the given text using Ollama.
// Returns a slice of floats representing the text in embedding space.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	req := api.EmbeddingRequest{
		Model:  e.model,
		Prompt: text,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP POST request to Ollama embeddings endpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Parse response
	var embResp struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(embResp.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return embResp.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts.
// Processes each text sequentially - for parallel processing, call Embed concurrently.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embed text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}
