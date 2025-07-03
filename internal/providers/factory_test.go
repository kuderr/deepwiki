package providers

import (
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/embedding"
	"github.com/kuderr/deepwiki/pkg/llm"
)

func TestNewLLMProvider(t *testing.T) {
	tests := []struct {
		name     string
		config   *llm.Config
		wantErr  bool
		errMsg   string
		wantType llm.ProviderType
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "OpenAI provider",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				APIKey:         "test-openai-key",
				Model:          "gpt-4o",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr:  false,
			wantType: llm.ProviderOpenAI,
		},
		{
			name: "Anthropic provider",
			config: &llm.Config{
				Provider:       llm.ProviderAnthropic,
				APIKey:         "test-anthropic-key",
				Model:          "claude-3-5-sonnet-20241022",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr:  false,
			wantType: llm.ProviderAnthropic,
		},
		{
			name: "unsupported provider",
			config: &llm.Config{
				Provider:       "unsupported",
				APIKey:         "test-key",
				Model:          "test-model",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr: true,
			errMsg:  "unsupported LLM provider",
		},
		{
			name: "OpenAI provider missing API key",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				Model:          "gpt-4o",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "Anthropic provider missing API key",
			config: &llm.Config{
				Provider:       llm.ProviderAnthropic,
				Model:          "claude-3-5-sonnet-20241022",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewLLMProvider(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLLMProvider() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewLLMProvider() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewLLMProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Error("NewLLMProvider() returned nil provider")
				return
			}

			// Test provider methods
			if provider.GetProviderType() != tt.wantType {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), tt.wantType)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}
		})
	}
}

func TestNewEmbeddingProvider(t *testing.T) {
	tests := []struct {
		name     string
		config   *embedding.Config
		wantErr  bool
		errMsg   string
		wantType embedding.ProviderType
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "OpenAI provider",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-openai-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr:  false,
			wantType: embedding.ProviderOpenAI,
		},
		{
			name: "Voyage provider",
			config: &embedding.Config{
				Provider:       embedding.ProviderVoyage,
				APIKey:         "test-voyage-key",
				Model:          "voyage-3-large",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1024,
			},
			wantErr:  false,
			wantType: embedding.ProviderVoyage,
		},
		{
			name: "Ollama provider",
			config: &embedding.Config{
				Provider:       embedding.ProviderOllama,
				Model:          "nomic-embed-text",
				BaseURL:        "http://localhost:11434",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     768,
			},
			wantErr:  false,
			wantType: embedding.ProviderOllama,
		},
		{
			name: "unsupported provider",
			config: &embedding.Config{
				Provider:       "unsupported",
				APIKey:         "test-key",
				Model:          "test-model",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
			},
			wantErr: true,
			errMsg:  "unsupported embedding provider",
		},
		{
			name: "OpenAI provider missing API key",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "Voyage provider missing API key",
			config: &embedding.Config{
				Provider:       embedding.ProviderVoyage,
				Model:          "voyage-3-large",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1024,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "Ollama provider missing base URL",
			config: &embedding.Config{
				Provider:       embedding.ProviderOllama,
				Model:          "nomic-embed-text",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     768,
			},
			wantErr: true,
			errMsg:  "base URL is required",
		},
		{
			name: "OpenAI provider missing model",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewEmbeddingProvider(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmbeddingProvider() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewEmbeddingProvider() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewEmbeddingProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Error("NewEmbeddingProvider() returned nil provider")
				return
			}

			// Test provider methods
			if provider.GetProviderType() != tt.wantType {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), tt.wantType)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}
		})
	}
}

func TestLLMProviderIntegration(t *testing.T) {
	// Test that we can create providers and they implement the correct interface
	config := &llm.Config{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "gpt-4o",
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   2.0,
	}

	provider, err := NewLLMProvider(config)
	if err != nil {
		t.Fatalf("NewLLMProvider() error = %v", err)
	}

	// Test that provider implements all required methods
	if provider.GetProviderType() != llm.ProviderOpenAI {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderOpenAI)
	}

	if provider.GetModel() != "gpt-4o" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "gpt-4o")
	}

	// Test token counting
	tokens, err := provider.CountTokens("hello world")
	if err != nil {
		t.Errorf("CountTokens() error = %v", err)
	}
	if tokens <= 0 {
		t.Errorf("CountTokens() = %d, want positive number", tokens)
	}

	// Test cost estimation
	cost := provider.EstimateCost(100, 50)
	if cost <= 0 {
		t.Errorf("EstimateCost() = %f, want positive number", cost)
	}

	// Test usage stats
	stats := provider.GetUsageStats()
	if stats.PromptTokens < 0 || stats.CompletionTokens < 0 || stats.TotalTokens < 0 {
		t.Errorf("GetUsageStats() returned negative values: %+v", stats)
	}

	// Test reset usage stats
	provider.ResetUsageStats()
	stats = provider.GetUsageStats()
	if stats.PromptTokens != 0 || stats.CompletionTokens != 0 || stats.TotalTokens != 0 || stats.EstimatedCost != 0 {
		t.Errorf("Usage stats should be zero after reset, got %+v", stats)
	}
}

func TestEmbeddingProviderIntegration(t *testing.T) {
	// Test that we can create providers and they implement the correct interface
	config := &embedding.Config{
		Provider:       embedding.ProviderOllama,
		Model:          "nomic-embed-text",
		BaseURL:        "http://localhost:11434",
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     768,
	}

	provider, err := NewEmbeddingProvider(config)
	if err != nil {
		t.Fatalf("NewEmbeddingProvider() error = %v", err)
	}

	// Test that provider implements all required methods
	if provider.GetProviderType() != embedding.ProviderOllama {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOllama)
	}

	if provider.GetModel() != "nomic-embed-text" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "nomic-embed-text")
	}

	if provider.GetDimensions() != 768 {
		t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), 768)
	}

	if provider.GetMaxTokens() <= 0 {
		t.Errorf("GetMaxTokens() = %v, want positive number", provider.GetMaxTokens())
	}

	// Test token estimation
	tokens := provider.EstimateTokens("hello world")
	if tokens <= 0 {
		t.Errorf("EstimateTokens() = %d, want positive number", tokens)
	}

	// Test text splitting
	chunks := provider.SplitTextForEmbedding("hello world test", 1)
	if len(chunks) <= 0 {
		t.Errorf("SplitTextForEmbedding() returned %d chunks, want positive number", len(chunks))
	}
}

func TestProviderConfigValidation(t *testing.T) {
	// Test that factory properly validates configurations
	tests := []struct {
		name       string
		configType string
		setupFunc  func() interface{}
		wantErr    bool
	}{
		{
			name:       "valid LLM config",
			configType: "llm",
			setupFunc: func() interface{} {
				return &llm.Config{
					Provider:       llm.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "gpt-4o",
					MaxTokens:      4000,
					Temperature:    0.1,
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   2.0,
				}
			},
			wantErr: false,
		},
		{
			name:       "invalid LLM config - negative max tokens",
			configType: "llm",
			setupFunc: func() interface{} {
				return &llm.Config{
					Provider:       llm.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "gpt-4o",
					MaxTokens:      -1,
					Temperature:    0.1,
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   2.0,
				}
			},
			wantErr: true,
		},
		{
			name:       "valid embedding config",
			configType: "embedding",
			setupFunc: func() interface{} {
				return &embedding.Config{
					Provider:       embedding.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "text-embedding-3-small",
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   10.0,
					Dimensions:     1536,
				}
			},
			wantErr: false,
		},
		{
			name:       "invalid embedding config - negative rate limit",
			configType: "embedding",
			setupFunc: func() interface{} {
				return &embedding.Config{
					Provider:       embedding.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "text-embedding-3-small",
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   -1.0,
					Dimensions:     1536,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupFunc()

			var err error
			if tt.configType == "llm" {
				_, err = NewLLMProvider(config.(*llm.Config))
			} else {
				_, err = NewEmbeddingProvider(config.(*embedding.Config))
			}

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
