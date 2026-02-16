package agent

import (
	"regexp"
	"strings"

	"github.com/openai/openai-go/v3"
)

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

var (
	thinkBlockRE         = regexp.MustCompile(`(?s)<think>.*?</think>`)
	orphanedThinkEndRE   = regexp.MustCompile(`(?s)^.*?</think>\s*`)
	orphanedThinkStartRE = regexp.MustCompile(`(?s)<think>.*$`)
	multiBlankLinesRE    = regexp.MustCompile(`\n\s*\n\s*\n+`)
)

// cleanThinkingTags removes <think> tags and their content from LLM responses.
func cleanThinkingTags(content string) string {
	content = thinkBlockRE.ReplaceAllString(content, "")
	content = orphanedThinkEndRE.ReplaceAllString(content, "")
	content = orphanedThinkStartRE.ReplaceAllString(content, "")
	content = multiBlankLinesRE.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}
