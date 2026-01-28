package aiclient

import (
	"context"
	"fmt"
	"encoding/json"
	"log"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type GeminiClient struct {
	client *openai.Client
}

func NewGemini(apiKey string) (*GeminiClient, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"),
	)
	return &GeminiClient{client: &client}, nil
}

// Ask implements the AIService interface
func (g *GeminiClient) Ask(ctx context.Context, history []ChatMessage) (string, error) {
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
		Model:    "gemini-3-flash-preview",
	})
	if err != nil {
		return "", fmt.Errorf("gemini api error: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		return chatCompletion.Choices[0].Message.Content, nil
	}

	return "Gemini returned no response", nil
}

// AskWithTools function for tool calls
func (g *GeminiClient) AskWithTools(ctx context.Context, context string, tools []ToolDefinition) (AIResponse, error) {
	// Implementation for tool calls can be added here
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
	chatCompletion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(context),
		},
		Model: "gemini-3-flash-preview",
		Tools: apiTools,
	})

	if err != nil {
		return AIResponse{}, fmt.Errorf("gemini api error: %w", err)
	}

	if len(chatCompletion.Choices) > 0 {
		choice := chatCompletion.Choices[0]

		// 1. Check if there are tool calls
		if len(choice.Message.ToolCalls) > 0 {
			// 2. Pre-allocate the slice for better performance
			myToolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))

			// 3. Loop and convert
			for _, tc := range choice.Message.ToolCalls {

				var argsMap map[string]any

				// Convert the string into a map
				err := json.Unmarshal([]byte(tc.Function.Arguments), &argsMap)
				if err != nil {
					// Handle cases where the AI hallucinates invalid JSON
					log.Printf("AI generated invalid JSON: %v", err)
					continue
				}
				myToolCalls = append(myToolCalls, ToolCall{
					Name:      tc.Function.Name,      // Access via .Function
					Arguments: argsMap, // Use the unmarshaled map
				})
			}
			return AIResponse{
				Content:   choice.Message.Content,
				ToolCalls: myToolCalls,
			}, nil
		}
		return AIResponse{
			Content:   choice.Message.Content,
			ToolCalls: nil,
		}, nil
	}

	return AIResponse{}, fmt.Errorf("Gemini returned no response")
}
