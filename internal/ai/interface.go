package ai

import (
	"context"

	"github.com/Shreehari-Acharya/vayuu/internal/memory"
)

type AIService interface {
	Ask(ctx context.Context, history []memory.Message) (string, error)
}