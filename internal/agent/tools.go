package agent

import (
	"sort"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

// toOpenAITool converts a Tool into an OpenAI tool definition.
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

// buildOpenAITools converts a tool map into a stable, ordered list of OpenAI tool definitions.
func buildOpenAITools(tools map[string]Tool) []openai.ChatCompletionToolUnionParam {
	result := make([]openai.ChatCompletionToolUnionParam, 0, len(tools))
	if len(tools) == 0 {
		return result
	}

	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		result = append(result, toOpenAITool(tools[name]))
	}

	return result
}
