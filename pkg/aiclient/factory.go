package aiclient

// Config holds configuration for creating an AI client.
type Config struct {
	UseGroq   bool
	GroqKey   string
	GeminiKey string
}

// New creates an appropriate AI client based on the configuration.
func New(cfg Config) (Client, error) {
	if cfg.UseGroq {
		return NewGroq(cfg.GroqKey)
	}
	return NewGemini(cfg.GeminiKey)
}
