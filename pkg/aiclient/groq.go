package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// GroqClient implements Client interface for Groq API.
type GroqClient struct {
	client *openai.Client
}

// NewGroq creates a new Groq client with the provided API key.
func NewGroq(apiKey string) (*GroqClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("groq api key is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.groq.com/openai/v1"),
	)
	return &GroqClient{client: &client}, nil
}

// Complete implements Client.Complete for Groq.
func (g *GroqClient) Complete(ctx context.Context, messages []Message) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context cannot be nil")
	}

	apiMessages := g.buildMessages(messages)

	completion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: apiMessages,
		Model:    "mixtral-8x7b-32768",
	})
	if err != nil {
		return "", fmt.Errorf("groq api error: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return completion.Choices[0].Message.Content, nil
}

// CompleteWithTools implements Client.CompleteWithTools for Groq.
func (g *GroqClient) CompleteWithTools(ctx context.Context, systemPrompt string, tools []ToolDefinition) (Response, error) {
	if ctx == nil {
		return Response{}, fmt.Errorf("context cannot be nil")
	}

	apiTools := g.buildTools(tools)

	completion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
		},
		Model: "mixtral-8x7b-32768",
		Tools: apiTools,
	})
	if err != nil {
		return Response{}, fmt.Errorf("groq api error: %w", err)
	}

	if len(completion.Choices) == 0 {
		return Response{}, fmt.Errorf("groq returned no choices")
	}

	return g.parseResponse(completion.Choices[0])
}

// buildMessages converts internal Message format to OpenAI API format.
func (g *GroqClient) buildMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	apiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)
	apiMessages = append(apiMessages, openai.SystemMessage(SystemPrompt))

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			apiMessages = append(apiMessages, openai.UserMessage(msg.Content))
		case "assistant":
			apiMessages = append(apiMessages, openai.AssistantMessage(msg.Content))
		}
	}

	return apiMessages
}

// buildTools converts internal ToolDefinition format to OpenAI API format.
func (g *GroqClient) buildTools(tools []ToolDefinition) []openai.ChatCompletionToolUnionParam {
	apiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		apiTools[i] = openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Type: "function",
				Function: shared.FunctionDefinitionParam{
					Name:        tool.Name,
					Parameters:  tool.InputSchema,
					Description: openai.String(tool.Description),
				},
			},
		}
	}
	return apiTools
}

// parseResponse converts API response to internal Response format.
func (g *GroqClient) parseResponse(choice openai.ChatCompletionChoice) (Response, error) {
	if len(choice.Message.ToolCalls) == 0 {
		return Response{
			Content:   choice.Message.Content,
			ToolCalls: nil,
		}, nil
	}

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			log.Printf("failed to unmarshal tool arguments: %v", err)
			continue
		}

		toolCalls = append(toolCalls, ToolCall{
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return Response{
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
	}, nil
}
