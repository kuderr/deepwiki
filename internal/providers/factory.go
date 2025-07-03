package providers

import (
	"fmt"

	"github.com/kuderr/deepwiki/pkg/embedding"
	"github.com/kuderr/deepwiki/pkg/embedding/ollama"
	"github.com/kuderr/deepwiki/pkg/embedding/openai"
	"github.com/kuderr/deepwiki/pkg/embedding/voyage"
	"github.com/kuderr/deepwiki/pkg/llm"
	llmanthropic "github.com/kuderr/deepwiki/pkg/llm/anthropic"
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
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

// NewEmbeddingProvider creates a new embedding provider based on the configuration
func NewEmbeddingProvider(config *embedding.Config) (embedding.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate config
	if config.RateLimitRPS < 0 {
		return nil, fmt.Errorf("rate limit RPS cannot be negative")
	}
	if config.MaxRetries < 0 {
		return nil, fmt.Errorf("max retries cannot be negative")
	}

	switch config.Provider {
	case embedding.ProviderOpenAI:
		return openai.NewProvider(config)
	case embedding.ProviderVoyage:
		return voyage.NewProvider(config)
	case embedding.ProviderOllama:
		return ollama.NewProvider(config)
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", config.Provider)
	}
}
