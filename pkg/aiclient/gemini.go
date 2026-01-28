package aiclient

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type Client struct {
	genaiClient *genai.Client
}

func NewGemini(apiKey string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}
	return &Client{genaiClient: client}, nil
}

// Ask implements the AIService interface
func (c *Client) Ask(ctx context.Context, history []ChatMessage) (string, error) {

	// Pre-allocate with exact capacity
	messages := make([]*genai.Content, 0, len(history))

	for i := range history {
		// Use index to avoid copying ChatMessage
		role := "user"
		if history[i].Role == "assistant" {
			role = "model"
		}
		messages = append(messages, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: history[i].Content},
			},
		})
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleModel),
	}
	result, err := c.genaiClient.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		messages,
		config,
	)

	if err != nil {
		return "", err
	}
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", result.Candidates[0].Content.Parts[0]), nil
	}
	return "I couldn't generate a response.", nil
}
