package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/openai/openai-go/v3"
)

// CreateAgent initializes a new Agent with the given configuration.
func CreateAgent(systemPrompt string, cfg *config.Config) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is empty")
	}

	agent := &Agent{
		client:       createLLMInstance(cfg),
		model:        cfg.Model,
		tools:        make(map[string]Tool),
		toolsDirty:   true,
		systemPrompt: systemPrompt,
		workDir:      cfg.AgentWorkDir,
		memoryWriter: NewFileMemoryWriter(cfg.AgentWorkDir),
		logf:         fmt.Printf,
	}

	return agent, nil
}

// RegisterTool adds a tool to the agent's available tools.
func (a *Agent) RegisterTool(tool Tool) error {
	if a.tools == nil {
		a.tools = make(map[string]Tool)
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Handler == nil {
		return fmt.Errorf("tool %s: handler cannot be nil", tool.Name)
	}

	if _, exists := a.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s: already registered", tool.Name)
	}

	a.tools[tool.Name] = tool
	a.toolsDirty = true
	a.toolsCache = nil
	return nil
}

// RunAgent executes the agent loop with the given user input.
// It iteratively calls the LLM, handles tool invocations, and returns the final response.
func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {
	logger := a.logger()
	logger("\n[%s] User: %s\n", time.Now().Format(clockLayout), userInput)

	// Initialize conversation with system and user messages
	messages := a.initializeMessages(userInput)

	// Run the agentic loop
	response, conversationHistory, err := a.executeAgentLoop(ctx, messages)
	if err != nil {
		return "", err
	}

	// Persist the conversation for future reference
	if a.memoryWriter != nil {
		if err := a.memoryWriter.Write(conversationHistory); err != nil {
			logger("Warning: failed to update memory: %s\n", err.Error())
		}
	}

	logger("[%s] Assistant: %s\n\n", time.Now().Format(clockLayout), response)
	return response, nil
}

// initializeMessages creates the initial message list with system and user messages.
func (a *Agent) initializeMessages(userInput string) []openai.ChatCompletionMessageParamUnion {
	return []openai.ChatCompletionMessageParamUnion{
		systemMsg(a.systemPrompt),
		userMsg(userInput),
	}
}

// executeAgentLoop runs the main agent loop, handling LLM responses and tool calls.
func (a *Agent) executeAgentLoop(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, []openai.ChatCompletionMessageParamUnion, error) {
	consecutiveErrors := 0
	iterations := 0
	logger := a.logger()

	for {
		// Check iteration limits and cancellation
		if err := ctx.Err(); err != nil {
			return "", messages, err
		}
		if iterations >= maxAgentIterations {
			return "", messages, fmt.Errorf("exceeded maximum agent iterations (%d)", maxAgentIterations)
		}
		if consecutiveErrors >= maxConsecutiveLLMErrors {
			return "", messages, fmt.Errorf("exceeded maximum consecutive LLM errors (%d)", maxConsecutiveLLMErrors)
		}
		iterations++

		// Request LLM response
		resp, err := a.requestCompletion(ctx, messages)
		if err != nil {
			consecutiveErrors++
			logger("LLM error (attempt %d/%d): %s\n", consecutiveErrors, maxConsecutiveLLMErrors, err.Error())
			continue
		}

		// Reset error counter on successful LLM call
		consecutiveErrors = 0

		llmMessage := resp.Choices[0].Message

		// Check if LLM wants to call tools
		if len(llmMessage.ToolCalls) > 0 {
			// Add assistant's tool call message to history
			messages = append(messages, assistantMsg(llmMessage))

			// Execute the tool calls and add results to history
			if err := a.executeTools(llmMessage, &messages); err != nil {
				return "", messages, err
			}

			// Continue loop for next LLM call
			continue
		}

		// LLM provided final response (no tool calls)
		finalResponse := cleanThinkingTags(llmMessage.Content)
		messages = append(messages, assistantMsg(llmMessage))

		return finalResponse, messages, nil
	}
}

// executeTools executes all tool calls from the LLM message and appends results to message history.
func (a *Agent) executeTools(llmMessage openai.ChatCompletionMessage, messages *[]openai.ChatCompletionMessageParamUnion) error {
	a.logger()("Executing %d tool(s)...\n", len(llmMessage.ToolCalls))

	for i, toolCall := range llmMessage.ToolCalls {
		if err := a.executeSingleTool(i, toolCall, messages); err != nil {
			return err
		}
	}

	return nil
}

// executeSingleTool executes a single tool call and records the result.
func (a *Agent) executeSingleTool(index int, toolCall openai.ChatCompletionMessageToolCallUnion, messages *[]openai.ChatCompletionMessageParamUnion) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %s panicked: %v", toolCall.Function.Name, r)
		}
	}()

	// Validate tool exists
	tool, ok := a.tools[toolCall.Function.Name]
	if !ok {
		return a.toolNotFoundError(toolCall.Function.Name)
	}

	// Parse tool arguments
	args := map[string]any{}
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			return fmt.Errorf("tool %s: failed to parse arguments: %w\nArguments: %s", toolCall.Function.Name, err, toolCall.Function.Arguments)
		}
	}

	// Execute the tool and measure elapsed time
	startTime := time.Now()
	result := tool.Handler(args)
	elapsed := time.Since(startTime)

	// Log execution details with result preview
	resultPreview := truncateString(result, resultPreviewLength, resultPreviewSuffix)
	a.logger()("   %d. %s (%v) --- %s\n", index+1, toolCall.Function.Name, elapsed, resultPreview)

	// Add tool result to message history
	*messages = append(*messages, toolCallMsg(toolCall.ID, result))

	return nil
}

// toolNotFoundError generates a helpful error message listing available tools.
func (a *Agent) toolNotFoundError(toolName string) error {
	availableTools := make([]string, 0, len(a.tools))
	for name := range a.tools {
		availableTools = append(availableTools, name)
	}
	return fmt.Errorf("unknown tool: %s (available: %v)", toolName, availableTools)
}

// truncateString truncates a string to a maximum length with an optional suffix.
func truncateString(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-len(suffix)] + suffix
}

// requestCompletion requests a completion from the LLM with the current message history.
func (a *Agent) requestCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       a.model,
		Messages:    messages,
		Tools:       a.openAITools(),
		Temperature: openai.Float(defaultTemperature),
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	return resp, nil
}

// openAITools returns a cached list of OpenAI tool definitions.
func (a *Agent) openAITools() []openai.ChatCompletionToolUnionParam {
	if !a.toolsDirty && a.toolsCache != nil {
		return a.toolsCache
	}

	a.toolsCache = buildOpenAITools(a.tools)
	a.toolsDirty = false
	return a.toolsCache
}

func (a *Agent) logger() Logf {
	if a.logf != nil {
		return a.logf
	}

	return fmt.Printf
}
