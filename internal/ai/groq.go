package ai

import (
	"context"
	"fmt"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type GroqClient struct {
	client *openai.Client
}

func NewGroq(apiKey string) (*GroqClient, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.groq.com/openai/v1"), // Note: correct Groq URL
	)
	return &GroqClient{client: &client}, nil
}

// Ask implements the AIService interface
func (g *GroqClient) Ask(ctx context.Context, prompt string) (string, error) {
	chatCompletion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
			openai.SystemMessage(systemPrompt),
		},
		Model: "openai/gpt-oss-120b",
	})
	if err != nil {
		return "", fmt.Errorf("groq api error: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return chatCompletion.Choices[0].Message.Content, nil
	}

	return "Groq returned no response", nil
}
