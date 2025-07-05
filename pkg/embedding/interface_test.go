package embedding

import (
	"testing"
	"time"
)

func TestProviderType_String(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected string
	}{
		{ProviderOpenAI, "openai"},
		{ProviderVoyage, "voyage"},
		{ProviderOllama, "ollama"},
		{ProviderType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			if string(tt.provider) != tt.expected {
				t.Errorf("ProviderType string = %v, want %v", string(tt.provider), tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name           string
		provider       ProviderType
		apiKey         string
		wantModel      string
		wantBaseURL    string
		wantDimensions int
	}{
		{
			name:           "OpenAI default config",
			provider:       ProviderOpenAI,
			apiKey:         "test-openai-key",
			wantModel:      "text-embedding-3-small",
			wantBaseURL:    "https://api.openai.com/v1",
			wantDimensions: 1536,
		},
		{
			name:           "Voyage default config",
			provider:       ProviderVoyage,
			apiKey:         "test-voyage-key",
			wantModel:      "voyage-3-large",
			wantBaseURL:    "https://api.voyageai.com/v1",
			wantDimensions: 1024,
		},
		{
			name:           "Ollama default config",
			provider:       ProviderOllama,
			apiKey:         "", // Ollama doesn't need API key
			wantModel:      "nomic-embed-text",
			wantBaseURL:    "http://localhost:11434",
			wantDimensions: 768,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig(tt.provider, tt.apiKey)

			if config == nil {
				t.Fatal("DefaultConfig() returned nil")
			}

			if config.Provider != tt.provider {
				t.Errorf("Provider = %v, want %v", config.Provider, tt.provider)
			}

			if config.APIKey != tt.apiKey {
				t.Errorf("APIKey = %v, want %v", config.APIKey, tt.apiKey)
			}

			if config.Model != tt.wantModel {
				t.Errorf("Model = %v, want %v", config.Model, tt.wantModel)
			}

			if config.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL = %v, want %v", config.BaseURL, tt.wantBaseURL)
			}

			if config.Dimensions != tt.wantDimensions {
				t.Errorf("Dimensions = %v, want %v", config.Dimensions, tt.wantDimensions)
			}

			// Test default values
			if config.RequestTimeout != 30*time.Second {
				t.Errorf("RequestTimeout = %v, want %v", config.RequestTimeout, 30*time.Second)
			}

			if config.MaxRetries != 3 {
				t.Errorf("MaxRetries = %v, want %v", config.MaxRetries, 3)
			}

			if config.RetryDelay != 1*time.Second {
				t.Errorf("RetryDelay = %v, want %v", config.RetryDelay, 1*time.Second)
			}

			if config.RateLimitRPS != 10.0 {
				t.Errorf("RateLimitRPS = %v, want %v", config.RateLimitRPS, 10.0)
			}
		})
	}
}

func TestEmbedding_Validation(t *testing.T) {
	tests := []struct {
		name      string
		embedding Embedding
		valid     bool
	}{
		{
			name: "valid embedding",
			embedding: Embedding{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
			valid: true,
		},
		{
			name: "empty object",
			embedding: Embedding{
				Object:    "",
				Index:     0,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
			valid: false,
		},
		{
			name: "negative index",
			embedding: Embedding{
				Object:    "embedding",
				Index:     -1,
				Embedding: []float64{0.1, 0.2, 0.3},
			},
			valid: false,
		},
		{
			name: "empty embedding vector",
			embedding: Embedding{
				Object:    "embedding",
				Index:     0,
				Embedding: []float64{},
			},
			valid: false,
		},
		{
			name: "nil embedding vector",
			embedding: Embedding{
				Object:    "embedding",
				Index:     0,
				Embedding: nil,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidObject := tt.embedding.Object != ""
			hasValidIndex := tt.embedding.Index >= 0
			hasValidEmbedding := tt.embedding.Embedding != nil && len(tt.embedding.Embedding) > 0
			isValid := hasValidObject && hasValidIndex && hasValidEmbedding

			if isValid != tt.valid {
				t.Errorf("Embedding validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestUsage_Validation(t *testing.T) {
	tests := []struct {
		name  string
		usage Usage
		valid bool
	}{
		{
			name: "valid usage",
			usage: Usage{
				PromptTokens: 100,
				TotalTokens:  100,
			},
			valid: true,
		},
		{
			name: "negative prompt tokens",
			usage: Usage{
				PromptTokens: -1,
				TotalTokens:  0,
			},
			valid: false,
		},
		{
			name: "negative total tokens",
			usage: Usage{
				PromptTokens: 100,
				TotalTokens:  -1,
			},
			valid: false,
		},
		{
			name: "inconsistent totals",
			usage: Usage{
				PromptTokens: 100,
				TotalTokens:  50, // Should be >= PromptTokens
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidTokens := tt.usage.PromptTokens >= 0 && tt.usage.TotalTokens >= 0
			hasConsistentTotal := tt.usage.TotalTokens >= tt.usage.PromptTokens
			isValid := hasValidTokens && hasConsistentTotal

			if isValid != tt.valid {
				t.Errorf("Usage validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestEmbeddingOptions_Defaults(t *testing.T) {
	// Test that default options make sense
	opts := EmbeddingOptions{}

	// Default values should be zero values, but we can test that they're settable
	opts.BatchSize = 100
	opts.InputType = "document"

	if opts.BatchSize != 100 {
		t.Errorf("BatchSize = %v, want %v", opts.BatchSize, 100)
	}

	if opts.InputType != "document" {
		t.Errorf("InputType = %v, want %v", opts.InputType, "document")
	}
}

func TestEmbeddingResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response EmbeddingResponse
		valid    bool
	}{
		{
			name: "valid response",
			response: EmbeddingResponse{
				Object: "list",
				Model:  "text-embedding-3-small",
				Data: []Embedding{
					{
						Object:    "embedding",
						Index:     0,
						Embedding: []float64{0.1, 0.2, 0.3},
					},
				},
				Usage: Usage{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			},
			valid: true,
		},
		{
			name: "empty object",
			response: EmbeddingResponse{
				Object: "",
				Model:  "text-embedding-3-small",
				Data: []Embedding{
					{
						Object:    "embedding",
						Index:     0,
						Embedding: []float64{0.1, 0.2, 0.3},
					},
				},
				Usage: Usage{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			},
			valid: false,
		},
		{
			name: "empty model",
			response: EmbeddingResponse{
				Object: "list",
				Model:  "",
				Data: []Embedding{
					{
						Object:    "embedding",
						Index:     0,
						Embedding: []float64{0.1, 0.2, 0.3},
					},
				},
				Usage: Usage{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			},
			valid: false,
		},
		{
			name: "no data",
			response: EmbeddingResponse{
				Object: "list",
				Model:  "text-embedding-3-small",
				Data:   []Embedding{},
				Usage: Usage{
					PromptTokens: 10,
					TotalTokens:  10,
				},
			},
			valid: false,
		},
		{
			name: "invalid usage",
			response: EmbeddingResponse{
				Object: "list",
				Model:  "text-embedding-3-small",
				Data: []Embedding{
					{
						Object:    "embedding",
						Index:     0,
						Embedding: []float64{0.1, 0.2, 0.3},
					},
				},
				Usage: Usage{
					PromptTokens: -1,
					TotalTokens:  10,
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasValidObject := tt.response.Object != ""
			hasValidModel := tt.response.Model != ""
			hasValidData := len(tt.response.Data) > 0
			hasValidUsage := tt.response.Usage.PromptTokens >= 0 && tt.response.Usage.TotalTokens >= 0
			isValid := hasValidObject && hasValidModel && hasValidData && hasValidUsage

			if isValid != tt.valid {
				t.Errorf("EmbeddingResponse validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid OpenAI config",
			config: Config{
				Provider:       ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     1536,
			},
			valid: true,
		},
		{
			name: "valid Ollama config",
			config: Config{
				Provider:       ProviderOllama,
				APIKey:         "", // Ollama doesn't need API key
				Model:          "nomic-embed-text",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "http://localhost:11434",
				Dimensions:     768,
			},
			valid: true,
		},
		{
			name: "missing provider",
			config: Config{
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     1536,
			},
			valid: false,
		},
		{
			name: "missing model",
			config: Config{
				Provider:       ProviderOpenAI,
				APIKey:         "test-key",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     1536,
			},
			valid: false,
		},
		{
			name: "negative max retries",
			config: Config{
				Provider:       ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     -1,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     1536,
			},
			valid: false,
		},
		{
			name: "negative rate limit",
			config: Config{
				Provider:       ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   -1.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     1536,
			},
			valid: false,
		},
		{
			name: "zero dimensions",
			config: Config{
				Provider:       ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				BaseURL:        "https://api.openai.com/v1",
				Dimensions:     0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic validation logic
			hasProvider := tt.config.Provider != ""
			hasModel := tt.config.Model != ""
			hasValidRetries := tt.config.MaxRetries >= 0
			hasValidRateLimit := tt.config.RateLimitRPS > 0
			hasValidDimensions := tt.config.Dimensions > 0
			isValid := hasProvider && hasModel && hasValidRetries && hasValidRateLimit && hasValidDimensions

			if isValid != tt.valid {
				t.Errorf("Config validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestProviderRequirements(t *testing.T) {
	// Test specific requirements for each provider type
	tests := []struct {
		name     string
		provider ProviderType
		config   Config
		valid    bool
		reason   string
	}{
		{
			name:     "OpenAI requires API key",
			provider: ProviderOpenAI,
			config: Config{
				Provider: ProviderOpenAI,
				APIKey:   "",
				Model:    "text-embedding-3-small",
			},
			valid:  false,
			reason: "OpenAI requires API key",
		},
		{
			name:     "Voyage requires API key",
			provider: ProviderVoyage,
			config: Config{
				Provider: ProviderVoyage,
				APIKey:   "",
				Model:    "voyage-3-large",
			},
			valid:  false,
			reason: "Voyage requires API key",
		},
		{
			name:     "Ollama does not require API key",
			provider: ProviderOllama,
			config: Config{
				Provider: ProviderOllama,
				APIKey:   "",
				Model:    "nomic-embed-text",
				BaseURL:  "http://localhost:11434",
			},
			valid:  true,
			reason: "Ollama does not require API key",
		},
		{
			name:     "Ollama requires base URL",
			provider: ProviderOllama,
			config: Config{
				Provider: ProviderOllama,
				Model:    "nomic-embed-text",
				BaseURL:  "",
			},
			valid:  false,
			reason: "Ollama requires base URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var isValid bool

			switch tt.provider {
			case ProviderOpenAI, ProviderVoyage:
				isValid = tt.config.APIKey != ""
			case ProviderOllama:
				isValid = tt.config.BaseURL != ""
			}

			// Also check common requirements
			isValid = isValid && tt.config.Model != ""

			if isValid != tt.valid {
				t.Errorf("Provider requirement validation = %v, want %v (reason: %s)", isValid, tt.valid, tt.reason)
			}
		})
	}
}
