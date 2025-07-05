package factory

import (
	"fmt"

	"github.com/kuderr/deepwiki/pkg/embedding"
	"github.com/kuderr/deepwiki/pkg/embedding/ollama"
	"github.com/kuderr/deepwiki/pkg/embedding/openai"
	"github.com/kuderr/deepwiki/pkg/embedding/voyage"
)

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
