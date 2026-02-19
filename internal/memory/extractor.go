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

type FactExtractor struct {
	client *openai.Client
	model  string
}

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

type ExtractedFact struct {
	Type     string `json:"type"` // "fact", "preference", "topic"
	Key      string `json:"key"`  // e.g., "name", "food", "hobby"
	Value    string `json:"value"`
	Category string `json:"category,omitempty"`
}

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
			systemMsg(`You extract structured facts from conversations. Always respond with valid JSON array.`),
			userMsg(prompt),
		},
		Temperature: openai.Float(0.3),
	})
	if err != nil {
		return nil, fmt.Errorf("LLM fact extraction failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

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

func systemMsg(text string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfSystem: &openai.ChatCompletionSystemMessageParam{
			Role: "system",
			Content: openai.ChatCompletionSystemMessageParamContentUnion{
				OfString: openai.String(text),
			},
		},
	}
}

func userMsg(text string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Role: "user",
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(text),
			},
		},
	}
}
