package factory

import (
	"strings"
	"testing"
	"time"

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
			name: "Ollama provider",
			config: &llm.Config{
				Provider:       llm.ProviderOllama,
				Model:          "llama3.1",
				BaseURL:        "http://localhost:11434",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   1.0,
			},
			wantErr:  false,
			wantType: llm.ProviderOllama,
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
		{
			name: "negative max tokens",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "gpt-4o",
				MaxTokens:      -1,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr: true,
			errMsg:  "max tokens cannot be negative",
		},
		{
			name: "negative rate limit",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "gpt-4o",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   -1.0,
			},
			wantErr: true,
			errMsg:  "rate limit RPS cannot be negative",
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
	if cost < 0 {
		t.Errorf("EstimateCost() = %f, want non-negative number", cost)
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

func TestOllamaProviderIntegration(t *testing.T) {
	// Test Ollama provider specifically
	config := &llm.Config{
		Provider:       llm.ProviderOllama,
		Model:          "llama3.1",
		BaseURL:        "http://localhost:11434",
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   1.0,
	}

	provider, err := NewLLMProvider(config)
	if err != nil {
		t.Fatalf("NewLLMProvider() error = %v", err)
	}

	// Test that provider implements all required methods
	if provider.GetProviderType() != llm.ProviderOllama {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderOllama)
	}

	if provider.GetModel() != "llama3.1" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "llama3.1")
	}

	// Test token counting
	tokens, err := provider.CountTokens("hello world")
	if err != nil {
		t.Errorf("CountTokens() error = %v", err)
	}
	if tokens <= 0 {
		t.Errorf("CountTokens() = %d, want positive number", tokens)
	}

	// Test cost estimation (should be 0 for Ollama)
	cost := provider.EstimateCost(100, 50)
	if cost != 0.0 {
		t.Errorf("EstimateCost() = %f, want 0.0 for Ollama", cost)
	}

	// Test usage stats
	stats := provider.GetUsageStats()
	if stats.PromptTokens < 0 || stats.CompletionTokens < 0 || stats.TotalTokens < 0 {
		t.Errorf("GetUsageStats() returned negative values: %+v", stats)
	}
}

func TestLLMProviderConfigValidation(t *testing.T) {
	// Test that factory properly validates configurations
	tests := []struct {
		name      string
		setupFunc func() *llm.Config
		wantErr   bool
	}{
		{
			name: "valid OpenAI config",
			setupFunc: func() *llm.Config {
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
			name: "valid Ollama config",
			setupFunc: func() *llm.Config {
				return &llm.Config{
					Provider:       llm.ProviderOllama,
					Model:          "llama3.1",
					BaseURL:        "http://localhost:11434",
					MaxTokens:      4000,
					Temperature:    0.1,
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   1.0,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid config - negative max tokens",
			setupFunc: func() *llm.Config {
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
			name: "invalid config - negative rate limit",
			setupFunc: func() *llm.Config {
				return &llm.Config{
					Provider:       llm.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "gpt-4o",
					MaxTokens:      4000,
					Temperature:    0.1,
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   -1.0,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupFunc()
			_, err := NewLLMProvider(config)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
