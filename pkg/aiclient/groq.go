package aiclient

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
		option.WithBaseURL("https://api.groq.com/openai/v1"),
	)
	return &GroqClient{client: &client}, nil
}

// Ask implements the AIService interface
func (g *GroqClient) Ask(ctx context.Context, history []ChatMessage) (string, error) {
	// Pre-allocate with exact capacity needed: system prompt + history
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
	messages = append(messages, openai.SystemMessage(systemPrompt))

	for i := range history {
		// Use index to avoid copying ChatMessage
		switch history[i].Role {
		case "user":
			messages = append(messages, openai.UserMessage(history[i].Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(history[i].Content))
		}
	}

	chatCompletion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    "openai/gpt-oss-120b",
	})
	if err != nil {
		return "", fmt.Errorf("groq api error: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return chatCompletion.Choices[0].Message.Content, nil
	}

	return "Groq returned no response", nil
}
