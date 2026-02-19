package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Shreehari-Acharya/vayuu/config"
	"github.com/Shreehari-Acharya/vayuu/internal/memory"
	"github.com/openai/openai-go/v3"
)

func CreateAgent(systemPrompt string, cfg *config.Config) (*Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	agent := &Agent{
		client:       createLLMInstance(cfg),
		model:        cfg.Model,
		tools:        make(map[string]Tool),
		toolsDirty:   true,
		systemPrompt: systemPrompt,
		workDir:      cfg.AgentWorkDir,
		memoryWriter: NewFileMemoryWriter(cfg.AgentWorkDir),
	}

	mgr, err := memory.NewMemoryManager(memory.DefaultConfig())
	if err != nil {
		slog.Warn("failed to initialize memory manager, continuing without it", "error", err)
	} else {
		agent.memoryMgr = mgr
		slog.Info("memory manager initialized")
	}

	return agent, nil
}

func (a *Agent) RegisterTool(tool Tool) error {
	if a.tools == nil {
		a.tools = make(map[string]Tool)
	}
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool %q: handler is required", tool.Name)
	}
	if _, exists := a.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q: already registered", tool.Name)
	}

	a.tools[tool.Name] = tool
	a.toolsDirty = true
	a.toolsCache = nil
	return nil
}

func (a *Agent) RunAgent(ctx context.Context, userInput string) (string, error) {
	slog.Info("agent invoked", "input_len", len(userInput))

	systemPrompt := a.systemPrompt

	if a.memoryMgr != nil {
		memContext, err := a.memoryMgr.GetContext(ctx, userInput, 500)
		if err != nil {
			slog.Warn("failed to get memory context", "error", err)
		} else if memContext != "" {
			systemPrompt = systemPrompt + "\n\n" + memContext
		}
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		systemMsg(systemPrompt),
		userMsg(userInput),
	}

	response, history, err := a.runLoop(ctx, messages)
	if err != nil {
		return "", err
	}

	if a.memoryWriter != nil {
		if err := a.memoryWriter.Write(history); err != nil {
			slog.Warn("failed to persist memory", "error", err)
		}
	}

	if a.memoryMgr != nil && response != "" {
		go a.extractAndStoreMemory(ctx, userInput, response)
	}

	slog.Info("agent completed", "response_len", len(response))
	return response, nil
}

func (a *Agent) extractAndStoreMemory(ctx context.Context, userInput, assistantResponse string) {
	if a.memoryMgr == nil {
		return
	}

	facts := extractFacts(assistantResponse)
	for _, fact := range facts {
		if err := a.memoryMgr.AddFact(ctx, fact, map[string]string{
			"source": "conversation",
		}); err != nil {
			slog.Warn("failed to store fact", "fact", fact, "error", err)
		}
	}
}

func extractFacts(text string) []string {
	var facts []string

	factIndicators := []string{
		"remember that", "i prefer", "i like", "i don't like",
		"my name is", "i am", "i live in", "i work as",
		"always", "never", "usually",
	}

	lower := text
	for _, indicator := range factIndicators {
		if idx := contains(lower, indicator); idx != -1 {
			sentence := extractSentence(lower, idx)
			if len(sentence) > 10 && len(sentence) < 200 {
				facts = append(facts, sentence)
			}
		}
	}

	return facts
}

func contains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if len(s) >= i+len(substr) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func extractSentence(s string, start int) string {
	end := start
	for end < len(s) && s[end] != '.' && s[end] != '\n' {
		end++
	}
	return strings.TrimSpace(s[start:end])
}

func (a *Agent) runLoop(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, []openai.ChatCompletionMessageParamUnion, error) {
	var consecutiveErrors, iterations int

	for {
		if err := ctx.Err(); err != nil {
			return "", messages, fmt.Errorf("context cancelled: %w", err)
		}
		if iterations >= maxAgentIterations {
			return "", messages, fmt.Errorf("exceeded max iterations (%d)", maxAgentIterations)
		}
		if consecutiveErrors >= maxConsecutiveLLMErrors {
			return "", messages, fmt.Errorf("exceeded max consecutive LLM errors (%d)", maxConsecutiveLLMErrors)
		}
		iterations++

		resp, err := a.requestCompletion(ctx, messages)
		if err != nil {
			consecutiveErrors++
			slog.Error("LLM request failed", "attempt", consecutiveErrors, "max", maxConsecutiveLLMErrors, "error", err)
			continue
		}
		consecutiveErrors = 0

		llmMsg := resp.Choices[0].Message

		if len(llmMsg.ToolCalls) == 0 {
			finalText := cleanThinkingTags(llmMsg.Content)
			messages = append(messages, assistantMsg(llmMsg))
			return finalText, messages, nil
		}

		messages = append(messages, assistantMsg(llmMsg))
		if err := a.dispatchToolCalls(llmMsg.ToolCalls, &messages); err != nil {
			return "", messages, err
		}
	}
}

func (a *Agent) dispatchToolCalls(calls []openai.ChatCompletionMessageToolCallUnion, messages *[]openai.ChatCompletionMessageParamUnion) error {
	slog.Info("dispatching tool calls", "count", len(calls))

	for i, call := range calls {
		result, err := a.invokeTool(call)
		if err != nil {
			return err
		}

		preview := result
		if len(preview) > resultPreviewLength {
			preview = preview[:resultPreviewLength] + resultPreviewSuffix
		}
		slog.Debug("tool result", "index", i+1, "name", call.Function.Name, "preview", preview)

		*messages = append(*messages, toolCallMsg(call.ID, result))
	}
	return nil
}

func (a *Agent) invokeTool(call openai.ChatCompletionMessageToolCallUnion) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %q panicked: %v", call.Function.Name, r)
		}
	}()

	tool, ok := a.tools[call.Function.Name]
	if !ok {
		available := make([]string, 0, len(a.tools))
		for name := range a.tools {
			available = append(available, name)
		}
		return "", fmt.Errorf("unknown tool %q (available: %v)", call.Function.Name, available)
	}

	args := map[string]any{}
	if call.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("tool %q: invalid arguments: %w", call.Function.Name, err)
		}
	}

	start := time.Now()
	result = tool.Handler(args)
	slog.Info("tool executed", "name", call.Function.Name, "duration", time.Since(start))

	return result, nil
}

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

func (a *Agent) openAITools() []openai.ChatCompletionToolUnionParam {
	if !a.toolsDirty && a.toolsCache != nil {
		return a.toolsCache
	}
	a.toolsCache = buildOpenAITools(a.tools)
	a.toolsDirty = false
	return a.toolsCache
}
