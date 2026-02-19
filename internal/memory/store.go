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

type VectorStore struct {
	baseURL    string
	client     *http.Client
	collection string
	vectorDim  int
}

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

func (vs *VectorStore) ensureCollection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", vs.baseURL+"/collections/"+vs.collection, nil)
	if err != nil {
		return err
	}

	resp, err := vs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	createReq := map[string]any{
		"name": vs.collection,
		"vectors": map[string]any{
			"size":     vs.vectorDim,
			"distance": "Cosine",
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

type qdrantPoint struct {
	ID      interface{}    `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

type upsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

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

type searchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type searchResponse struct {
	Result []struct {
		ID      interface{}    `json:"id"`
		Score   float64        `json:"score"`
		Payload map[string]any `json:"payload"`
	} `json:"result"`
}

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

func (vs *VectorStore) Close() error {
	return nil
}

func parseTime(s string) interface{} {
	return s
}
