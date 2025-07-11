package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/embedding"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *embedding.Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "missing API key",
			config: &embedding.Config{
				Provider: embedding.ProviderOpenAI,
				Model:    "text-embedding-3-small",
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "missing model",
			config: &embedding.Config{
				Provider: embedding.ProviderOpenAI,
				APIKey:   "test-key",
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "valid config",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr: false,
		},
		{
			name: "config with custom base URL",
			config: &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "text-embedding-3-small",
				BaseURL:        "https://custom.openai.com/v1",
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProvider() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewProvider() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Error("NewProvider() returned nil provider")
				return
			}

			// Test provider methods
			if provider.GetProviderType() != embedding.ProviderOpenAI {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOpenAI)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}

			if provider.GetDimensions() != tt.config.Dimensions {
				t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), tt.config.Dimensions)
			}

			// Test default base URL is set
			if tt.config.BaseURL == "" && provider.(*OpenAIProvider).config.BaseURL != "https://api.openai.com/v1" {
				t.Errorf("Default BaseURL not set correctly")
			}
		})
	}
}

func TestOpenAIProvider_CreateEmbeddings(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected path /embeddings, got %s", r.URL.Path)
		}

		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Parse request
		var request EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if request.Model != "text-embedding-3-small" {
			t.Errorf("Expected model 'text-embedding-3-small', got %s", request.Model)
		}

		if len(request.Input) != 2 {
			t.Errorf("Expected 2 input texts, got %d", len(request.Input))
		}

		// Mock response
		response := EmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-3-small",
			Data: []EmbeddingData{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: []float64{0.1, 0.2, 0.3, 0.4},
				},
				{
					Object:    "embedding",
					Index:     1,
					Embedding: []float64{0.5, 0.6, 0.7, 0.8},
				},
			},
			Usage: EmbeddingUsage{
				PromptTokens: 10,
				TotalTokens:  10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "text-embedding-3-small",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     1536,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	texts := []string{"hello world", "test text"}

	response, err := provider.CreateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("CreateEmbeddings() error = %v", err)
	}

	// Verify response
	if response == nil {
		t.Fatal("CreateEmbeddings() returned nil response")
	}

	if response.Object != "list" {
		t.Errorf("Expected object 'list', got %s", response.Object)
	}

	if response.Model != "text-embedding-3-small" {
		t.Errorf("Expected model 'text-embedding-3-small', got %s", response.Model)
	}

	if len(response.Data) != 2 {
		t.Fatalf("Expected 2 embeddings, got %d", len(response.Data))
	}

	// Verify first embedding
	embedding1 := response.Data[0]
	if embedding1.Object != "embedding" {
		t.Errorf("Expected object 'embedding', got %s", embedding1.Object)
	}
	if embedding1.Index != 0 {
		t.Errorf("Expected index 0, got %d", embedding1.Index)
	}
	if len(embedding1.Embedding) != 4 {
		t.Errorf("Expected 4 dimensions, got %d", len(embedding1.Embedding))
	}
	if embedding1.Embedding[0] != 0.1 || embedding1.Embedding[1] != 0.2 {
		t.Errorf("Unexpected embedding values: %v", embedding1.Embedding)
	}

	// Verify second embedding
	embedding2 := response.Data[1]
	if embedding2.Index != 1 {
		t.Errorf("Expected index 1, got %d", embedding2.Index)
	}
	if len(embedding2.Embedding) != 4 {
		t.Errorf("Expected 4 dimensions, got %d", len(embedding2.Embedding))
	}
	if embedding2.Embedding[0] != 0.5 || embedding2.Embedding[1] != 0.6 {
		t.Errorf("Unexpected embedding values: %v", embedding2.Embedding)
	}

	// Verify usage
	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", response.Usage.PromptTokens)
	}
	if response.Usage.TotalTokens != 10 {
		t.Errorf("Expected total tokens 10, got %d", response.Usage.TotalTokens)
	}
}

func TestOpenAIProvider_CreateEmbeddingsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to verify options
		var request EmbeddingRequest
		json.NewDecoder(r.Body).Decode(&request)

		if request.Dimensions == nil || *request.Dimensions != 512 {
			t.Errorf("Expected dimensions 512, got %v", request.Dimensions)
		}

		response := EmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-3-small",
			Data: []EmbeddingData{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: make([]float64, 512), // 512 dimensions as requested
				},
			},
			Usage: EmbeddingUsage{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "text-embedding-3-small",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     1536,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	texts := []string{"test"}

	options := embedding.EmbeddingOptions{
		BatchSize: 5,
		InputType: "document",
	}

	// Update provider config to use 512 dimensions for this test
	provider.(*OpenAIProvider).config.Dimensions = 512

	response, err := provider.CreateEmbeddings(ctx, texts, options)
	if err != nil {
		t.Fatalf("CreateEmbeddings() error = %v", err)
	}

	if response == nil {
		t.Fatal("CreateEmbeddings() returned nil response")
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 embedding, got %d", len(response.Data))
	}

	if len(response.Data[0].Embedding) != 512 {
		t.Errorf("Expected 512 dimensions, got %d", len(response.Data[0].Embedding))
	}
}

func TestOpenAIProvider_CreateEmbeddingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid API key",
				Type:    "invalid_request_error",
				Code:    "invalid_api_key",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOpenAI,
		APIKey:         "invalid-key",
		Model:          "text-embedding-3-small",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     0, // No retries for this test
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     1536,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	texts := []string{"test"}

	_, err = provider.CreateEmbeddings(ctx, texts)
	if err == nil {
		t.Fatal("Expected error from CreateEmbeddings(), got nil")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error to contain 'Invalid API key', got: %v", err)
	}
}

func TestOpenAIProvider_CreateEmbeddingsRetry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			// Fail first two requests
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on third request
		response := EmbeddingResponse{
			Object: "list",
			Model:  "text-embedding-3-small",
			Data: []EmbeddingData{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: []float64{0.1, 0.2, 0.3, 0.4},
				},
			},
			Usage: EmbeddingUsage{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "text-embedding-3-small",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     10 * time.Millisecond, // Short delay for testing
		RateLimitRPS:   10.0,
		Dimensions:     1536,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	texts := []string{"test"}

	response, err := provider.CreateEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("CreateEmbeddings() error = %v", err)
	}

	if response == nil {
		t.Fatal("CreateEmbeddings() returned nil response")
	}

	if requestCount != 3 {
		t.Errorf("Expected 3 requests (2 failures + 1 success), got %d", requestCount)
	}
}

func TestOpenAIProvider_EstimateTokens(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"this is a longer text with more words", 9},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			count := provider.EstimateTokens(tt.text)
			if count != tt.expected {
				t.Errorf("EstimateTokens() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestOpenAIProvider_SplitTextForEmbedding(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	tests := []struct {
		name      string
		text      string
		maxTokens int
		expected  int // number of chunks
	}{
		{
			name:      "short text",
			text:      "hello world",
			maxTokens: 10,
			expected:  1,
		},
		{
			name:      "text that needs splitting",
			text:      "this is a very long text that should be split into multiple chunks when the token limit is reached",
			maxTokens: 5,
			expected:  5, // Approximately 19 words (~24 tokens) / 5 = 5 chunks
		},
		{
			name:      "empty text",
			text:      "",
			maxTokens: 10,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := provider.SplitTextForEmbedding(tt.text, tt.maxTokens)
			if len(chunks) != tt.expected {
				t.Errorf("SplitTextForEmbedding() returned %d chunks, want %d", len(chunks), tt.expected)
			}

			// Verify each chunk is within token limit
			for i, chunk := range chunks {
				tokens := provider.EstimateTokens(chunk)
				if tokens > tt.maxTokens {
					t.Errorf("Chunk %d has %d tokens, exceeds limit of %d", i, tokens, tt.maxTokens)
				}
			}

			// Verify all chunks together contain the original text
			if tt.text != "" && len(chunks) > 0 {
				combined := strings.Join(chunks, " ")
				// The combined text should contain all the original words
				originalWords := strings.Fields(tt.text)
				combinedWords := strings.Fields(combined)
				if len(originalWords) != len(combinedWords) {
					t.Errorf(
						"Original text had %d words, combined chunks have %d words",
						len(originalWords),
						len(combinedWords),
					)
				}
			}
		})
	}
}

func TestOpenAIProvider_GetMaxTokens(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"text-embedding-ada-002", 8191},
		{"text-embedding-3-small", 8191},
		{"text-embedding-3-large", 8191},
		{"unknown-model", 8191}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			config := &embedding.Config{
				Provider:       embedding.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          tt.model,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   10.0,
				Dimensions:     1536,
			}

			provider, err := NewProvider(config)
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}

			maxTokens := provider.GetMaxTokens()
			if maxTokens != tt.expected {
				t.Errorf("GetMaxTokens() = %d, want %d", maxTokens, tt.expected)
			}
		})
	}
}

func TestOpenAIProvider_ProviderInfo(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Test provider type
	if provider.GetProviderType() != embedding.ProviderOpenAI {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOpenAI)
	}

	// Test model name
	if provider.GetModel() != "text-embedding-3-small" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "text-embedding-3-small")
	}

	// Test dimensions
	if provider.GetDimensions() != 1536 {
		t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), 1536)
	}

	// Test max tokens
	if provider.GetMaxTokens() != 8191 {
		t.Errorf("GetMaxTokens() = %v, want %v", provider.GetMaxTokens(), 8191)
	}
}

func TestOpenAIProvider_EmptyTextList(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	texts := []string{}

	_, err = provider.CreateEmbeddings(ctx, texts)
	if err == nil {
		t.Fatal("Expected error for empty text list, got nil")
	}

	if !strings.Contains(err.Error(), "no texts provided") {
		t.Errorf("Expected error to contain 'no texts provided', got: %v", err)
	}
}
