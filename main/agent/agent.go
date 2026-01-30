package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/main/prompts"
	"github.com/openai/openai-go/v3"
)

type Agent struct {
	client   *openai.Client
	model    string
	tools    map[string]Tool
	messages []openai.ChatCompletionMessageParamUnion
}

func NewAgent(systemPrompt string, cfg *config.Config) *Agent {

	llm := createLLMInstance(cfg)

	return &Agent{
		client: llm,
		model:  cfg.Model,
		tools:  make(map[string]Tool),
		messages: []openai.ChatCompletionMessageParamUnion{
			systemMsg(prompts.GetSystemPrompt()),
		},
	}
}

func (a *Agent) RegisterTool(tool Tool) {
	if a.tools == nil {
		a.tools = make(map[string]Tool)
	}
	fmt.Println("Registering tool:", tool.Name)
	a.tools[tool.Name] = tool
}

func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {
	fmt.Printf("%s Received: %s\n", time.Now().Format("2006-01-02 15:04:05"), userInput)
	a.messages = append(a.messages, userMsg(userInput))

	for {
		// Trim messages to fit within limits
		a.messages = trimMessages(a.messages)

		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: a.messages,
			Tools:    buildOpenAITools(a.tools),
		})
		if err != nil {
			return "", err
		}

		msg := resp.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			a.messages = append(a.messages, assistantMsgFromResponse(msg))

			for _, tc := range msg.ToolCalls {
				tool, ok := a.tools[tc.Function.Name]
				if !ok {
					return "", fmt.Errorf("unknown tool: %s", tc.Function.Name)
				}

				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					return "", err
				}

				result := tool.Handler(args)
				// Append tool result back
				a.messages = append(a.messages, toolCallMsg(
					tc.ID,
					result,
				))
			}
			// continue loop -> send tool result back to LLM
			continue
		}

		// Final response
		if msg.Content != "" {
			a.messages = append(a.messages, assistantMsgFromResponse(msg))
			return msg.Content, nil
		}

		return "", fmt.Errorf("agent stopped with no output")
	}
}

func trimMessages(msgs []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {

	const maxMessages = 20
	const maxArgLength = 20
	const preserveLastN = 2

	if len(msgs) <= 1 {
		return msgs
	}
	// Always keep the first message (system prompt)
	system := systemMsg(prompts.GetSystemPrompt())

	// 1. Process all messages to shrink massive tool arguments
	for i := 0; i < len(msgs)-preserveLastN; i++ {
		// We only care about Assistant messages that contain ToolCalls
		if msgs[i].OfAssistant != nil && len(msgs[i].OfAssistant.ToolCalls) > 0 {
			for j := range msgs[i].OfAssistant.ToolCalls {
				tc := &msgs[i].OfAssistant.ToolCalls[j]

				// truncate all the parameters, 20chars followed by ...
				jsonMap := make(map[string]any)
				if err := json.Unmarshal([]byte(tc.OfFunction.Function.Arguments), &jsonMap); err == nil {
					for key, val := range jsonMap {
						valStr := fmt.Sprintf("%v", val)
						if len(valStr) > maxArgLength {
							jsonMap[key] = valStr[:maxArgLength] + "..."
						} else {
							jsonMap[key] = valStr
						}
					}
					// Marshal back to JSON string
					if truncatedArgs, err := json.Marshal(jsonMap); err == nil {
						tc.OfFunction.Function.Arguments = string(truncatedArgs)
					}
				}
			}
		}
	}
	// logging history
	 fmt.Print("\033[H\033[2J") 
	if b, err := json.Marshal(msgs); err == nil {
		fmt.Printf("[%s] Current History: %s\n", time.Now().Format("15:04:05"), string(b))
	}

	// if we have not reached maxMessages, return as is
	if len(msgs) <= maxMessages {
		return msgs
	}
	// Keep last (maxMessages - 1) messages
	recent := msgs[len(msgs)-(maxMessages-1):]

	// Rebuild slice
	trimmed := make([]openai.ChatCompletionMessageParamUnion, 0, maxMessages)
	trimmed = append(trimmed, system)
	trimmed = append(trimmed, recent...)

	return trimmed
}
