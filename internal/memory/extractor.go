package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// FactExtractor uses an LLM to extract structured facts from conversations.
// It analyzes conversation text and returns facts, preferences, and topics.
type FactExtractor struct {
	client *openai.Client // LLM client
	model  string         // Model name for extraction
}

// NewFactExtractor creates a new FactExtractor using the provided config.
func NewFactExtractor(cfg *config.Config) *FactExtractor {
	client := openai.NewClient(
		option.WithAPIKey(cfg.ApiKey),
		option.WithBaseURL(cfg.ApiBaseURL),
	)
	return &FactExtractor{
		client: &client,
		model:  cfg.Model,
	}
}

// ExtractedFact represents a structured fact extracted from conversation.
// Type can be: "fact", "preference", or "topic"
type ExtractedFact struct {
	Type     string `json:"type"`               // Type of fact
	Key      string `json:"key"`                // Identifier (e.g., "name", "food")
	Value    string `json:"value"`              // The extracted value
	Category string `json:"category,omitempty"` // Category for preferences
}

// ExtractFacts analyzes a conversation and extracts structured information about the user.
// Returns a slice of ExtractedFact or an error if extraction fails.
func (e *FactExtractor) ExtractFacts(ctx context.Context, conversation string) ([]ExtractedFact, error) {
	prompt := `Analyze the following conversation and extract structured information about the user.

Extract facts in this JSON format:
[
  {"type": "fact", "key": "key_name", "value": "the fact"},
  {"type": "preference", "key": "preference_name", "value": "the preference", "category": "category"},
  {"type": "topic", "key": "topic_name", "value": "topic"}
]

Rules:
- Only extract if there's clear new information about the user
- "key" should be lowercase snake_case
- "category" for preferences: food, hobby, work, communication, other
- Return empty array if nothing significant to extract
- Keep values concise (under 50 words)

Conversation:
` + conversation

	resp, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: e.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			newSystemMsg(`You extract structured facts from conversations. Always respond with valid JSON array.`),
			newUserMsg(prompt),
		},
		Temperature: openai.Float(0.3),
	})
	if err != nil {
		return nil, fmt.Errorf("LLM fact extraction failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse response, stripping any markdown code blocks
	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var facts []ExtractedFact
	if err := json.Unmarshal([]byte(content), &facts); err != nil {
		return nil, fmt.Errorf("parse facts: %w", err)
	}

	return facts, nil
}

// newSystemMsg creates a system message for the LLM.
func newSystemMsg(text string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfSystem: &openai.ChatCompletionSystemMessageParam{
			Role: "system",
			Content: openai.ChatCompletionSystemMessageParamContentUnion{
				OfString: openai.String(text),
			},
		},
	}
}

// newUserMsg creates a user message for the LLM.
func newUserMsg(text string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Role: "user",
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(text),
			},
		},
	}
}
