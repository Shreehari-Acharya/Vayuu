package main

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

func assistantMsg(text string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			Role: "assistant",
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(text),
			},
		},
	}
}

func toolCallMsg(result string) openai.ChatCompletionMessageParamUnion {
	return openai.ChatCompletionMessageParamUnion{
		OfTool: &openai.ChatCompletionToolMessageParam{
			Role: "tool",
			Content: openai.ChatCompletionToolMessageParamContentUnion{
				OfString: openai.String(result),
			},
		},
	}
}

