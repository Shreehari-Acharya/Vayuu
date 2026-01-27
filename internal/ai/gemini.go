package ai

import (
	"context"
	"fmt"

	"github.com/Shreehari-Acharya/vayuu/internal/memory"
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

func (c *Client) Ask(ctx context.Context, userQuery string, history []memory.Message) (string, error) {

	var messages []*genai.Content

	for _, msg := range history {
        role := "user"
        if msg.Role == "assistant" {
            role = "model"
        }
        messages = append(messages, &genai.Content{
            Role: role,
            Parts: []*genai.Part{
                {Text: msg.Content},
            },
        })
    }

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
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
