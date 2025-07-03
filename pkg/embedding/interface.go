package embedding

import (
	"context"
	"time"
)

// ProviderType represents the type of embedding provider
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderVoyage ProviderType = "voyage"
	ProviderOllama ProviderType = "ollama"
)

// EmbeddingResponse represents a response from embeddings API
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Usage represents usage information
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// EmbeddingOptions holds options for embedding requests
type EmbeddingOptions struct {
	BatchSize int
	InputType string // "query" or "document" for some providers
}

// Provider interface defines the embedding provider methods
type Provider interface {
	// Create embeddings for given texts
	CreateEmbeddings(ctx context.Context, texts []string, opts ...EmbeddingOptions) (*EmbeddingResponse, error)

	// Provider info
	GetProviderType() ProviderType
	GetModel() string
	GetDimensions() int
	GetMaxTokens() int

	// Token estimation
	EstimateTokens(text string) int

	// Text processing helpers
	SplitTextForEmbedding(text string, maxTokens int) []string
}

// Config represents embedding provider configuration
type Config struct {
	Provider       ProviderType  `yaml:"provider"`
	APIKey         string        `yaml:"api_key,omitempty"`
	Model          string        `yaml:"model"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	RateLimitRPS   float64       `yaml:"rate_limit_rps"`

	// Provider-specific configurations
	BaseURL    string `yaml:"base_url,omitempty"`   // For custom endpoints (Ollama)
	Dimensions int    `yaml:"dimensions,omitempty"` // For some providers
}

// DefaultConfig returns default configuration for the specified provider
func DefaultConfig(provider ProviderType, apiKey string) *Config {
	base := &Config{
		Provider:       provider,
		APIKey:         apiKey,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
	}

	switch provider {
	case ProviderOpenAI:
		base.Model = "text-embedding-3-small"
		base.BaseURL = "https://api.openai.com/v1"
		base.Dimensions = 1536
	case ProviderVoyage:
		base.Model = "voyage-3-large"
		base.BaseURL = "https://api.voyageai.com/v1"
		base.Dimensions = 1024
	case ProviderOllama:
		base.Model = "nomic-embed-text"
		base.BaseURL = "http://localhost:11434"
		base.Dimensions = 768
		base.APIKey = "" // Ollama doesn't need API key
	}

	return base
}
