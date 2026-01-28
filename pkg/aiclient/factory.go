package aiclient

import "fmt"

// NewAIService creates an AI service based on the provider type
func NewAIService(useGroq bool, groqKey, geminiKey string) (AIService, error) {
	if useGroq {
		if groqKey == "" {
			return nil, fmt.Errorf("groq key is required")
		}
		return NewGroq(groqKey)
	}
	if geminiKey == "" {
		return nil, fmt.Errorf("gemini key is required")
	}
	return NewGemini(geminiKey)
}
