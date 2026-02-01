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
	client *openai.Client
	model  string
	tools  map[string]Tool
}

func NewAgent(systemPrompt string, cfg *config.Config) *Agent {
	llm := createLLMInstance(cfg)

	return &Agent{
		client: llm,
		model:  cfg.Model,
		tools:  make(map[string]Tool),
	}
}

func (a *Agent) RegisterTool(tool Tool) {
	if a.tools == nil {
		a.tools = make(map[string]Tool)
	}
	a.tools[tool.Name] = tool
}

const (
	maxLLMErrors        = 3
	maxIterations       = 20
	temperatureLow float64 = 0.2
)

func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {
	fmt.Printf("\n[%s] User: %s\n", time.Now().Format("15:04:05"), userInput)

	messages := []openai.ChatCompletionMessageParamUnion{
		systemMsg(prompts.SystemPrompt),
		userMsg(userInput),
	}

	errCount := 0
	iterations := 0
	var finalResponse string

	for {
		// Safety checks
		if errCount >= maxLLMErrors {
			return "", fmt.Errorf("exceeded maximum LLM errors (%d)", maxLLMErrors)
		}
		if iterations >= maxIterations {
			return "", fmt.Errorf("exceeded maximum iterations (%d)", maxIterations)
		}
		iterations++

		// Get LLM response
		resp, err := a.getLLMResponse(ctx, messages)
		if err != nil {
			fmt.Printf("LLM Error (attempt %d/%d): %s\n", errCount+1, maxLLMErrors, err.Error())
			errCount++
			continue
		}
		errCount = 0

		msg := resp.Choices[0].Message

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			messages = append(messages, assistantMsgFromResponse(msg))

			if err := a.handleToolCalls(msg, &messages); err != nil {
				return "", err
			}
			continue
		}

		// Final response - this is what gets returned to user
		finalResponse = cleanThinkingTags(msg.Content)
		messages = append(messages, assistantMsgFromResponse(msg))
		break
	}

	// Update memory with the complete conversation
	if err := updateMemoryFile(messages); err != nil {
		fmt.Printf("Memory update failed: %s\n", err.Error())
	}

	fmt.Printf("[%s] Assistant: %s\n\n", time.Now().Format("15:04:05"), finalResponse)
	
	return finalResponse, nil
}

func (a *Agent) getLLMResponse(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	return a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       a.model,
		Messages:    messages,
		Tools:       buildOpenAITools(a.tools),
		Temperature: openai.Float(temperatureLow),
	})
}

func (a *Agent) handleToolCalls(msg openai.ChatCompletionMessage, messages *[]openai.ChatCompletionMessageParamUnion) error {
	fmt.Printf("Executing %d tool(s)...\n", len(msg.ToolCalls))

	for i, tc := range msg.ToolCalls {
		tool, ok := a.tools[tc.Function.Name]
		if !ok {
			return fmt.Errorf("unknown tool: %s", tc.Function.Name)
		}

		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return fmt.Errorf("failed to parse arguments for %s: %w", tc.Function.Name, err)
		}

		startTime := time.Now()
		result := tool.Handler(args)
		elapsed := time.Since(startTime)

		preview := result
		if len(result) > 50 {
			preview = result[:47] + "..."
		}
		fmt.Printf("   %d. %s (%v) --- %s\n", i+1, tc.Function.Name, elapsed, preview)

		*messages = append(*messages, toolCallMsg(tc.ID, result))
	}

	return nil
}