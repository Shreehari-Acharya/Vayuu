package agent

// Agent configuration constants.
const (
	maxConsecutiveLLMErrors = 3
	maxAgentIterations      = 20
	defaultTemperature      = 0.2
	resultPreviewLength     = 50
	resultPreviewSuffix     = "..."
	defaultMemoryMaxSize    = 10 * 1024 * 1024 // 10MB
	memoryDirName           = "memory"
	dayFileLayout           = "2006-01-02"
	clockLayout             = "15:04:05"
)
