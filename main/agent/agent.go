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
	fmt.Println("Registering tool:", tool.Name)
	a.tools[tool.Name] = tool
}

const (
	maxLLMErrors     = 3
	maxIterations    = 20 // Prevent infinite loops
	resultPreviewLen = 100
)

func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {
	fmt.Printf("%s Received: %s\n", time.Now().Format("2006-01-02 15:04:05"), userInput)
	
	messages := []openai.ChatCompletionMessageParamUnion{
		systemMsg(prompts.SystemPrompt),
		userMsg(userInput),
	}

	errCount := 0
	iterations := 0

	for {
		// Safety checks
		if errCount >= maxLLMErrors {
			return "", fmt.Errorf("exceeded maximum LLM errors (%d), aborting", maxLLMErrors)
		}
		if iterations >= maxIterations {
			return "", fmt.Errorf("exceeded maximum iterations (%d), possible infinite loop", maxIterations)
		}
		iterations++

		// Get LLM response
		resp, err := a.getLLMResponse(ctx, messages)
		if err != nil {
			fmt.Printf("Error from OpenAI (attempt %d/%d): %s\n", errCount+1, maxLLMErrors, err.Error())
			errCount++
			continue
		}
		errCount = 0 // Reset on success

		msg := resp.Choices[0].Message

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			messages = append(messages, assistantMsgFromResponse(msg))
			
			if err := a.handleToolCalls(msg, &messages); err != nil {
				return "", err
			}
			continue
		}

		// Final response
		if msg.Content != "" {
			return msg.Content, nil
		}

		return "", fmt.Errorf("agent stopped with no output")
	}
}

func (a *Agent) getLLMResponse(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	return a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    a.model,
		Messages: messages,
		Tools:    buildOpenAITools(a.tools),
	})
}

func (a *Agent) handleToolCalls(msg openai.ChatCompletionMessage, messages *[]openai.ChatCompletionMessageParamUnion) error {
	if msg.Content != "" {
		fmt.Printf("Agent reasoning: %s\n", msg.Content)
	}

	for _, tc := range msg.ToolCalls {
		tool, ok := a.tools[tc.Function.Name]
		if !ok {
			return fmt.Errorf("unknown tool: %s", tc.Function.Name)
		}

		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return fmt.Errorf("failed to parse arguments for %s: %w", tc.Function.Name, err)
		}

		result := tool.Handler(args)
		
		preview := result
		if len(result) > resultPreviewLen {
			preview = result[:resultPreviewLen] + "..."
		}
		fmt.Printf("Executed tool '%s', result: %s\n", tc.Function.Name, preview)

		*messages = append(*messages, toolCallMsg(tc.ID, result))
	}
	
	return nil
}