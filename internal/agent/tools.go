package agent

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type ToolFunc func(args map[string]any) string

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema
	Handler     ToolFunc
}

func toOpenAITool(t Tool) openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{

			Function: shared.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  t.Parameters,
			},
		},
	}
}

func buildOpenAITools(tools map[string]Tool) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))

	for _, t := range tools {
		result = append(result, toOpenAITool(t))
	}

	return result
}
