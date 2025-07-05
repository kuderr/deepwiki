package config

import (
	"fmt"
	"time"

	"github.com/kuderr/deepwiki/pkg/embedding"
	"github.com/kuderr/deepwiki/pkg/llm"
)

// ProviderConfig contains configuration for both LLM and embedding providers
type ProviderConfig struct {
	LLM       LLMConfig       `yaml:"llm"`
	Embedding EmbeddingConfig `yaml:"embedding"`
}

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	Provider       string  `yaml:"provider"` // "openai", "anthropic", or "ollama"
	APIKey         string  `yaml:"api_key"`
	Model          string  `yaml:"model"`
	MaxTokens      int     `yaml:"max_tokens"`
	Temperature    float64 `yaml:"temperature"`
	RequestTimeout string  `yaml:"request_timeout"` // Duration string like "3m"
	MaxRetries     int     `yaml:"max_retries"`
	RetryDelay     string  `yaml:"retry_delay"` // Duration string like "1s"
	RateLimitRPS   float64 `yaml:"rate_limit_rps"`
	BaseURL        string  `yaml:"base_url"` // For custom endpoints
}

// EmbeddingConfig contains embedding provider configuration
type EmbeddingConfig struct {
	Provider       string  `yaml:"provider"` // "openai", "voyage", or "ollama"
	APIKey         string  `yaml:"api_key"`
	Model          string  `yaml:"model"`
	RequestTimeout string  `yaml:"request_timeout"` // Duration string like "30s"
	MaxRetries     int     `yaml:"max_retries"`
	RetryDelay     string  `yaml:"retry_delay"` // Duration string like "1s"
	RateLimitRPS   float64 `yaml:"rate_limit_rps"`
	BaseURL        string  `yaml:"base_url"`   // For custom endpoints (Ollama)
	Dimensions     int     `yaml:"dimensions"` // For some providers
}

// DefaultProviderConfig returns default provider configuration
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		LLM: LLMConfig{
			Provider:       "openai",
			Model:          "gpt-4o",
			MaxTokens:      4000,
			Temperature:    0.1,
			RequestTimeout: "3m",
			MaxRetries:     3,
			RetryDelay:     "1s",
			RateLimitRPS:   2.0,
		},
		Embedding: EmbeddingConfig{
			Provider:       "openai",
			Model:          "text-embedding-3-small",
			RequestTimeout: "30s",
			MaxRetries:     3,
			RetryDelay:     "1s",
			RateLimitRPS:   10.0,
		},
	}
}

// ToLLMConfig converts the application LLM config to llm.Config
func (c *LLMConfig) ToLLMConfig() (*llm.Config, error) {
	// Parse durations
	requestTimeout, err := time.ParseDuration(c.RequestTimeout)
	if err != nil {
		requestTimeout = 3 * time.Minute // Default
	}

	retryDelay, err := time.ParseDuration(c.RetryDelay)
	if err != nil {
		retryDelay = 1 * time.Second // Default
	}

	// Convert provider type
	var providerType llm.ProviderType
	switch c.Provider {
	case "openai":
		providerType = llm.ProviderOpenAI
	case "anthropic":
		providerType = llm.ProviderAnthropic
	case "ollama":
		providerType = llm.ProviderOllama
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", c.Provider)
	}

	config := &llm.Config{
		Provider:       providerType,
		APIKey:         c.APIKey,
		Model:          c.Model,
		MaxTokens:      c.MaxTokens,
		Temperature:    c.Temperature,
		RequestTimeout: requestTimeout,
		MaxRetries:     c.MaxRetries,
		RetryDelay:     retryDelay,
		RateLimitRPS:   c.RateLimitRPS,
		BaseURL:        c.BaseURL,
	}

	// Set defaults if not specified
	if config.Model == "" {
		switch providerType {
		case llm.ProviderOpenAI:
			config.Model = "gpt-4o"
		case llm.ProviderAnthropic:
			config.Model = "claude-3-5-sonnet-20241022"
		case llm.ProviderOllama:
			config.Model = "llama3.1"
		}
	}

	if config.BaseURL == "" {
		switch providerType {
		case llm.ProviderOpenAI:
			config.BaseURL = "https://api.openai.com/v1"
		case llm.ProviderAnthropic:
			config.BaseURL = "https://api.anthropic.com/v1"
		case llm.ProviderOllama:
			config.BaseURL = "http://localhost:11434"
		}
	}

	return config, nil
}

// ToEmbeddingConfig converts the application embedding config to embedding.Config
func (c *EmbeddingConfig) ToEmbeddingConfig() (*embedding.Config, error) {
	// Parse durations
	requestTimeout, err := time.ParseDuration(c.RequestTimeout)
	if err != nil {
		requestTimeout = 30 * time.Second // Default
	}

	retryDelay, err := time.ParseDuration(c.RetryDelay)
	if err != nil {
		retryDelay = 1 * time.Second // Default
	}

	// Convert provider type
	var providerType embedding.ProviderType
	switch c.Provider {
	case "openai":
		providerType = embedding.ProviderOpenAI
	case "voyage":
		providerType = embedding.ProviderVoyage
	case "ollama":
		providerType = embedding.ProviderOllama
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", c.Provider)
	}

	config := &embedding.Config{
		Provider:       providerType,
		APIKey:         c.APIKey,
		Model:          c.Model,
		RequestTimeout: requestTimeout,
		MaxRetries:     c.MaxRetries,
		RetryDelay:     retryDelay,
		RateLimitRPS:   c.RateLimitRPS,
		BaseURL:        c.BaseURL,
		Dimensions:     c.Dimensions,
	}

	// Set defaults if not specified
	if config.Model == "" {
		switch providerType {
		case embedding.ProviderOpenAI:
			config.Model = "text-embedding-3-small"
		case embedding.ProviderVoyage:
			config.Model = "voyage-3-large"
		case embedding.ProviderOllama:
			config.Model = "nomic-embed-text"
		}
	}

	if config.BaseURL == "" {
		switch providerType {
		case embedding.ProviderOpenAI:
			config.BaseURL = "https://api.openai.com/v1"
		case embedding.ProviderVoyage:
			config.BaseURL = "https://api.voyageai.com/v1"
		case embedding.ProviderOllama:
			config.BaseURL = "http://localhost:11434"
		}
	}

	return config, nil
}
