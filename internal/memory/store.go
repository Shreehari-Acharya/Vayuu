package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VectorStore provides a client for Qdrant vector database via REST API.
// It handles storing and searching vector embeddings for semantic memory.
type VectorStore struct {
	baseURL    string       // Qdrant REST API URL (http://localhost:6333)
	client     *http.Client // HTTP client with timeout
	collection string       // Name of the collection
	vectorDim  int          // Dimension of vectors (e.g., 768 for nomic-embed-text)
}

// NewVectorStore creates a new VectorStore and ensures the collection exists.
// Returns an error if Qdrant is not reachable or collection creation fails.
func NewVectorStore(cfg *Config) (*VectorStore, error) {
	vs := &VectorStore{
		baseURL:    "http://localhost:6333",
		client:     &http.Client{Timeout: 30 * time.Second},
		collection: cfg.CollectionName,
		vectorDim:  cfg.VectorDim,
	}

	if err := vs.ensureCollection(context.Background()); err != nil {
		return nil, err
	}

	return vs, nil
}

// ensureCollection checks if the collection exists, creating it if necessary.
// Uses Cosine distance for similarity matching.
func (vs *VectorStore) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	req, err := http.NewRequestWithContext(ctx, "GET", vs.baseURL+"/collections/"+vs.collection, nil)
	if err != nil {
		return err
	}

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Collection exists, nothing to do
	if resp.StatusCode == 200 {
		return nil
	}

	// Create collection
	createReq := map[string]any{
		"name": vs.collection,
		"vectors": map[string]any{
			"size":     vs.vectorDim,
			"distance": "Cosine", // Cosine similarity for semantic search
		},
	}

	body, _ := json.Marshal(createReq)
	req, err = http.NewRequestWithContext(ctx, "PUT", vs.baseURL+"/collections/"+vs.collection, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create collection failed: %s", string(b))
	}

	return nil
}

// qdrantPoint represents a single point in Qdrant (ID + vector + payload)
type qdrantPoint struct {
	ID      interface{}    `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

// upsertRequest is the request body for Qdrant's upsert endpoint
type upsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

// Upsert stores or updates a memory in the vector database.
// The ID should be unique - using UUID is recommended.
func (vs *VectorStore) Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error {
	point := qdrantPoint{
		ID:      id,
		Vector:  vector,
		Payload: payload,
	}

	reqBody := upsertRequest{Points: []qdrantPoint{point}}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		vs.baseURL+"/collections/"+vs.collection+"/points/upsert", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed: %s", string(b))
	}

	return nil
}

// searchRequest is the request body for Qdrant's search endpoint
type searchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

// searchResponse parses Qdrant's search response
type searchResponse struct {
	Result []struct {
		ID      interface{}    `json:"id"`
		Score   float64        `json:"score"`
		Payload map[string]any `json:"payload"`
	} `json:"result"`
}

// Search finds the most similar memories to the given vector.
// Returns results sorted by similarity score (highest first).
func (vs *VectorStore) Search(ctx context.Context, vector []float32, limit int, filter interface{}) ([]SearchResult, error) {
	reqBody := searchRequest{
		Vector:      vector,
		Limit:       limit,
		WithPayload: true,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		vs.baseURL+"/collections/"+vs.collection+"/points/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(b))
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert Qdrant results to our SearchResult type
	searchResults := make([]SearchResult, 0, len(result.Result))
	for _, r := range result.Result {
		content := ""
		memType := MemoryTypeFact
		createdAt := ""

		if p := r.Payload; p != nil {
			if c, ok := p["content"].(string); ok {
				content = c
			}
			if t, ok := p["type"].(string); ok {
				memType = MemoryType(t)
			}
			if ct, ok := p["created_at"].(string); ok {
				createdAt = ct
			}
		}

		idStr := fmt.Sprintf("%v", r.ID)
		searchResults = append(searchResults, SearchResult{
			Memory: Memory{
				ID:        idStr,
				Content:   content,
				Type:      memType,
				CreatedAt: createdAt,
			},
			Score: r.Score,
		})
	}

	return searchResults, nil
}

// Delete removes a memory by ID from the vector store.
func (vs *VectorStore) Delete(ctx context.Context, id string) error {
	reqBody := map[string][]string{"points": {id}}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		vs.baseURL+"/collections/"+vs.collection+"/points/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Close implements the io.Closer interface.
// Currently a no-op since HTTP client handles its own resources.
func (vs *VectorStore) Close() error {
	return nil
}
