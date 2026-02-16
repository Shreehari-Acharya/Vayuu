package agent

import (
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// createLLMInstance creates a new OpenAI client using the provided configuration.
func createLLMInstance(cfg *config.Config) *openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.ApiKey),
		option.WithBaseURL(cfg.ApiBaseURL),
	)

	return &client
}

// systemMsg creates a system role message.
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

// userMsg creates a user role message.
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

// assistantMsg creates an assistant role message from an LLM response.
func assistantMsg(msg openai.ChatCompletionMessage) openai.ChatCompletionMessageParamUnion {
	assistant := &openai.ChatCompletionAssistantMessageParam{
		Role: "assistant",
	}

	// Content may be empty when tool calls are present
	if msg.Content != "" {
		assistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
			OfString: openai.String(msg.Content),
		}
	}

	// Preserve tool calls if they exist
	if len(msg.ToolCalls) > 0 {
		assistant.ToolCalls = toolCallsToParams(msg.ToolCalls)
	}

	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: assistant,
	}
}

// toolCallMsg creates a tool role message to send tool output back to the model.
func toolCallMsg(toolCallID string, content string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCallID,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	}
}
