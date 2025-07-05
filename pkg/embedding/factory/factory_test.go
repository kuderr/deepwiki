package factory

import (
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/embedding"
)

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
		{
			name: "negative rate limit",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   -1.0,
				Dimensions:     1536,
			},
			wantErr: true,
			errMsg:  "rate limit RPS cannot be negative",
		},
		{
			name: "negative max retries",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     -1,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr: true,
			errMsg:  "max retries cannot be negative",
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

func TestOpenAIEmbeddingProviderIntegration(t *testing.T) {
	// Test OpenAI provider specifically
	config := &embedding.Config{
		Provider:       embedding.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "text-embedding-3-small",
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     1536,
	}

	provider, err := NewEmbeddingProvider(config)
	if err != nil {
		t.Fatalf("NewEmbeddingProvider() error = %v", err)
	}

	// Test that provider implements all required methods
	if provider.GetProviderType() != embedding.ProviderOpenAI {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOpenAI)
	}

	if provider.GetModel() != "text-embedding-3-small" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "text-embedding-3-small")
	}

	if provider.GetDimensions() != 1536 {
		t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), 1536)
	}
}

func TestVoyageEmbeddingProviderIntegration(t *testing.T) {
	// Test Voyage provider specifically
	config := &embedding.Config{
		Provider:       embedding.ProviderVoyage,
		APIKey:         "test-voyage-key",
		Model:          "voyage-3-large",
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     1024,
	}

	provider, err := NewEmbeddingProvider(config)
	if err != nil {
		t.Fatalf("NewEmbeddingProvider() error = %v", err)
	}

	// Test that provider implements all required methods
	if provider.GetProviderType() != embedding.ProviderVoyage {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderVoyage)
	}

	if provider.GetModel() != "voyage-3-large" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "voyage-3-large")
	}

	if provider.GetDimensions() != 1024 {
		t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), 1024)
	}
}

func TestEmbeddingProviderConfigValidation(t *testing.T) {
	// Test that factory properly validates configurations
	tests := []struct {
		name      string
		setupFunc func() *embedding.Config
		wantErr   bool
	}{
		{
			name: "valid OpenAI config",
			setupFunc: func() *embedding.Config {
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
			name: "valid Ollama config",
			setupFunc: func() *embedding.Config {
				return &embedding.Config{
					Provider:       embedding.ProviderOllama,
					Model:          "nomic-embed-text",
					BaseURL:        "http://localhost:11434",
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   10.0,
					Dimensions:     768,
				}
			},
			wantErr: false,
		},
		{
			name: "valid Voyage config",
			setupFunc: func() *embedding.Config {
				return &embedding.Config{
					Provider:       embedding.ProviderVoyage,
					APIKey:         "test-voyage-key",
					Model:          "voyage-3-large",
					RequestTimeout: 30 * time.Second,
					MaxRetries:     3,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   10.0,
					Dimensions:     1024,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid config - negative rate limit",
			setupFunc: func() *embedding.Config {
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
		{
			name: "invalid config - negative max retries",
			setupFunc: func() *embedding.Config {
				return &embedding.Config{
					Provider:       embedding.ProviderOpenAI,
					APIKey:         "test-key",
					Model:          "text-embedding-3-small",
					RequestTimeout: 30 * time.Second,
					MaxRetries:     -1,
					RetryDelay:     1 * time.Second,
					RateLimitRPS:   10.0,
					Dimensions:     1536,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupFunc()
			_, err := NewEmbeddingProvider(config)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
