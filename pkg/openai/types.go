package openai

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completion API
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatCompletionResponse represents a response from the chat completion API
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

// EmbeddingRequest represents a request to the embeddings API
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents a response from the embeddings API
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

// APIError represents an OpenAI API error
type APIError struct {
	ErrorInfo struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (e APIError) Error() string {
	return e.ErrorInfo.Message
}

// Config holds OpenAI client configuration
type Config struct {
	APIKey         string
	Model          string
	EmbeddingModel string
	MaxTokens      int
	Temperature    float64
	RequestTimeout time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	RateLimitRPS   float64
}

// DefaultConfig returns default OpenAI configuration
func DefaultConfig(apiKey string) *Config {
	return &Config{
		APIKey:         apiKey,
		Model:          "gpt-4o",
		EmbeddingModel: "text-embedding-3-small",
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 3 * 60 * time.Second, // 3 minutes. TODO: move to config
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   2.0,
	}
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

// EmbeddingOptions holds options for embedding requests
type EmbeddingOptions struct {
	BatchSize int
}

// Client interface defines the OpenAI client methods
type Client interface {
	// Chat completion methods
	ChatCompletion(ctx context.Context, messages []Message, opts ...ChatCompletionOptions) (*ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, messages []Message, handler StreamHandler, opts ...ChatCompletionOptions) error

	// Embedding methods
	CreateEmbeddings(ctx context.Context, texts []string, opts ...EmbeddingOptions) (*EmbeddingResponse, error)

	// Token counting methods
	CountTokens(text string) (int, error)
	EstimateCost(promptTokens, completionTokens int) float64

	// Rate limiting and stats
	GetUsageStats() TokenCount
	ResetUsageStats()
}
