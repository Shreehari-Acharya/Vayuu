package memory

import (
	"regexp"
	"strings"
)

var (
	thinkBlockRE         = regexp.MustCompile(`(?s)<think>.*?</think>`)
	orphanedThinkEndRE   = regexp.MustCompile(`(?s)^.*?</think>\s*`)
	orphanedThinkStartRE = regexp.MustCompile(`(?s)<think>.*$`)
	multiBlankLinesRE    = regexp.MustCompile(`\n\s*\n\s*\n+`)
)

// CleanThinkingTags removes <think> tags and their content from LLM responses.
func CleanThinkingTags(content string) string {
	content = thinkBlockRE.ReplaceAllString(content, "")
	content = orphanedThinkEndRE.ReplaceAllString(content, "")
	content = orphanedThinkStartRE.ReplaceAllString(content, "")
	content = multiBlankLinesRE.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}
