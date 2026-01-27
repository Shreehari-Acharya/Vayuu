package ai

import "context"

type AIService interface {
	Ask(ctx context.Context, prompt string) (string, error)
}