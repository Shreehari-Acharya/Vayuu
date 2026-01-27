package ai

import (
	"context"
	"fmt"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
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
func (g *GroqClient) Ask(ctx context.Context, nextQuery string, history []memory.Message) (string, error) {
	
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	for _, msg := range history {
        switch msg.Role {
		case "user":
            messages = append(messages, openai.UserMessage(msg.Content))
        case "assistant":
            messages = append(messages, openai.AssistantMessage(msg.Content))
        }
    }

	chatCompletion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
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
