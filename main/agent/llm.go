package agent

import (
	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func createLLMInstance(cfg *config.Config) *openai.Client {

	apiKey := cfg.ApiKey
	apiBaseURL := cfg.ApiBaseURL

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiBaseURL),
	)

	return &client
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

func assistantMsgFromResponse(msg openai.ChatCompletionMessage) openai.ChatCompletionMessageParamUnion {
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

func toolCallMsg(toolCallId string, content string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role:       "tool",
			ToolCallID: toolCallId,
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(content),
			},
		},
	}
}

func toolCallsToParams(
	calls []openai.ChatCompletionMessageToolCallUnion,
) []openai.ChatCompletionMessageToolCallUnionParam {

	out := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(calls))

	for _, c := range calls {

		out = append(out, openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
				ID: c.ID,
				Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
					Name:      c.Function.Name,
					Arguments: c.Function.Arguments,
				},
			},
		})
	}

	return out
}
