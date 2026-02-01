package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"regexp"
	"strings"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var cfig *config.Config

func createLLMInstance(cfg *config.Config) *openai.Client {
	apiKey := cfg.ApiKey
	apiBaseURL := cfg.ApiBaseURL

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(apiBaseURL),
	)

	cfig = cfg
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

type MemoryEntry struct {
	Timestamp string `json:"timestamp"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

// cleanThinkingTags removes thinking tags and their content from LLM responses
func cleanThinkingTags(content string) string {
	// Step 1: Remove complete thinking blocks: <think>...</think>
	re1 := regexp.MustCompile(`(?s)<think>.*?</think>`)
	content = re1.ReplaceAllString(content, "")
	
	// Step 2: Handle orphaned closing tag </think> by removing everything before it
	// This assumes the opening <think> was truncated/missing
	re2 := regexp.MustCompile(`(?s)^.*?</think>\s*`)
	content = re2.ReplaceAllString(content, "")
	
	// Step 3: Handle orphaned opening tag <think> by removing everything after it
	// This assumes the closing </think> was truncated/missing
	re3 := regexp.MustCompile(`(?s)<think>.*$`)
	content = re3.ReplaceAllString(content, "")
	
	// Step 4: Clean up extra whitespace
	re4 := regexp.MustCompile(`\n\s*\n\s*\n+`)
	content = re4.ReplaceAllString(content, "\n\n")
	
	return strings.TrimSpace(content)
}

func updateMemoryFile(messages []openai.ChatCompletionMessageParamUnion) error {
	memoryDir := cfig.AgentWorkDir + "/memory"
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/%s.jsonl", memoryDir, time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	timestamp := time.Now().Format("15:04:05")

	// Only save user and assistant messages (skip system and tool messages)
	for _, msg := range messages {
		var entry MemoryEntry
		entry.Timestamp = timestamp

		if msg.OfUser != nil {
			entry.Role = "user"
			if msg.OfUser.Content.OfString.String() != "" {
				entry.Content = msg.OfUser.Content.OfString.Value
			}
		} else if msg.OfAssistant != nil {
			entry.Role = "assistant"
			
			if msg.OfAssistant.Content.OfString.String() != "" {
				entry.Content = cleanThinkingTags(msg.OfAssistant.Content.OfString.Value)
			} else {
				continue // Skip if no content
			}
		} else {
			continue 
		}

		// Only write entries with content
		if entry.Content != "" {
			if err := encoder.Encode(entry); err != nil {
				return err
			}
		}
	}

	return nil
}