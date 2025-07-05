package factory

import (
	"fmt"

	"github.com/kuderr/deepwiki/pkg/llm"
	llmanthropic "github.com/kuderr/deepwiki/pkg/llm/anthropic"
	llmollama "github.com/kuderr/deepwiki/pkg/llm/ollama"
	llmopenai "github.com/kuderr/deepwiki/pkg/llm/openai"
)

// NewLLMProvider creates a new LLM provider based on the configuration
func NewLLMProvider(config *llm.Config) (llm.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate config
	if config.MaxTokens < 0 {
		return nil, fmt.Errorf("max tokens cannot be negative")
	}
	if config.RateLimitRPS < 0 {
		return nil, fmt.Errorf("rate limit RPS cannot be negative")
	}

	switch config.Provider {
	case llm.ProviderOpenAI:
		return llmopenai.NewProvider(config)
	case llm.ProviderAnthropic:
		return llmanthropic.NewProvider(config)
	case llm.ProviderOllama:
		return llmollama.NewProvider(config)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}
