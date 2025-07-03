package llm

import (
	"context"
	"time"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents a response from chat completion API
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a streaming choice
type StreamChoice struct {
	Index int     `json:"index"`
	Delta Message `json:"delta"`
}

// TokenCount represents token usage statistics
type TokenCount struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	EstimatedCost    float64
}

// StreamHandler is a function type for handling streaming responses
type StreamHandler func(chunk StreamResponse) error

// ChatCompletionOptions holds options for chat completion requests
type ChatCompletionOptions struct {
	MaxTokens   int
	Temperature float64
	Stream      bool
	OnStream    StreamHandler
}

// Provider interface defines the LLM provider methods
type Provider interface {
	// Chat completion methods
	ChatCompletion(
		ctx context.Context,
		messages []Message,
		opts ...ChatCompletionOptions,
	) (*ChatCompletionResponse, error)

	ChatCompletionStream(
		ctx context.Context,
		messages []Message,
		handler StreamHandler,
		opts ...ChatCompletionOptions,
	) error

	// Token counting methods
	CountTokens(text string) (int, error)
	EstimateCost(promptTokens, completionTokens int) float64

	// Usage tracking
	GetUsageStats() TokenCount
	ResetUsageStats()

	// Provider info
	GetProviderType() ProviderType
	GetModel() string
}

// Config represents LLM provider configuration
type Config struct {
	Provider       ProviderType  `yaml:"provider"`
	APIKey         string        `yaml:"api_key"`
	Model          string        `yaml:"model"`
	MaxTokens      int           `yaml:"max_tokens"`
	Temperature    float64       `yaml:"temperature"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	RateLimitRPS   float64       `yaml:"rate_limit_rps"`

	// Provider-specific configurations
	BaseURL string `yaml:"base_url,omitempty"` // For custom endpoints
}

// DefaultConfig returns default configuration for the specified provider
func DefaultConfig(provider ProviderType, apiKey string) *Config {
	base := &Config{
		Provider:       provider,
		APIKey:         apiKey,
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 3 * time.Minute,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   2.0,
	}

	switch provider {
	case ProviderOpenAI:
		base.Model = "gpt-4o"
		base.BaseURL = "https://api.openai.com/v1"
	case ProviderAnthropic:
		base.Model = "claude-3-5-sonnet-20241022"
		base.BaseURL = "https://api.anthropic.com/v1"
	}

	return base
}
