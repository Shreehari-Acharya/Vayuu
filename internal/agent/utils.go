package agent

import (
	"github.com/openai/openai-go/v3"
)

// toolCallsToParams converts OpenAI tool call structs to the parameter format
// expected by the API for sending tool results back to the model.
func toolCallsToParams(calls []openai.ChatCompletionMessageToolCallUnion) []openai.ChatCompletionMessageToolCallUnionParam {
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
