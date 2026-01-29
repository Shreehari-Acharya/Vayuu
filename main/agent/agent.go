package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"

)

type Agent struct {
	client  *openai.Client
	model   string
	tools   map[string]Tool
	messages []openai.ChatCompletionMessageParamUnion
}

func NewAgent(systemPrompt string, cfg *config.Config) *Agent {

	llm := createLLMInstance(cfg)

	return &Agent{
		client: llm,
		model: cfg.Model,
		tools: make(map[string]Tool),
		messages: []openai.ChatCompletionMessageParamUnion{
			systemMsg(systemPrompt),
		},
	}
}

func (a *Agent) RegisterTool(tool Tool) {
	if a.tools == nil {
		a.tools = make(map[string]Tool)
	}
	a.tools[tool.Name] = tool
}

func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {

	a.messages = append(a.messages, userMsg(userInput))
	a.messages = trimMessages(a.messages)

	for {
		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: a.messages,
			Tools: buildOpenAITools(a.tools),
			MaxTokens: openai.Int(1000),
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
			a.messages = trimMessages(a.messages)
			return msg.Content, nil
		}

		return "", fmt.Errorf("agent stopped with no output")
	}
}



func trimMessages(msgs []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	
	const maxMessages = 30

	if len(msgs) <= maxMessages {
		return msgs
	}

	// Always keep the first message (system prompt)
	system := msgs[0]

	// Keep last (maxMessages - 1) messages
	recent := msgs[len(msgs)-(maxMessages-1):]

	// Rebuild slice
	trimmed := make([]openai.ChatCompletionMessageParamUnion, 0, maxMessages)
	trimmed = append(trimmed, system)
	trimmed = append(trimmed, recent...)

	return trimmed
}