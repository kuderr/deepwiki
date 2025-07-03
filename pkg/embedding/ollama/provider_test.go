package ollama

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
			name: "missing base URL",
			config: &embedding.Config{
				Provider: embedding.ProviderOllama,
				Model:    "nomic-embed-text",
			},
			wantErr: true,
			errMsg:  "base URL is required for Ollama provider",
		},
		{
			name: "missing model",
			config: &embedding.Config{
				Provider: embedding.ProviderOllama,
				BaseURL:  "http://localhost:11434",
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "valid config",
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
			if provider.GetProviderType() != embedding.ProviderOllama {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOllama)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}

			if provider.GetDimensions() != tt.config.Dimensions {
				t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), tt.config.Dimensions)
			}

			expectedMaxTokens := 8192 // nomic-embed-text supports 8192 tokens
			if provider.GetMaxTokens() != expectedMaxTokens {
				t.Errorf("GetMaxTokens() = %v, want %v", provider.GetMaxTokens(), expectedMaxTokens)
			}
		})
	}
}

func TestOllamaProvider_CreateEmbeddings(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected path /api/embeddings, got %s", r.URL.Path)
		}

		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Parse request
		var request OllamaEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if request.Model != "nomic-embed-text" {
			t.Errorf("Expected model 'nomic-embed-text', got %s", request.Model)
		}

		// Mock response with different embeddings for each text
		var response OllamaEmbedResponse
		if request.Prompt == "hello world" {
			response = OllamaEmbedResponse{
				Embedding: []float64{0.1, 0.2, 0.3},
			}
		} else if request.Prompt == "test text" {
			response = OllamaEmbedResponse{
				Embedding: []float64{0.4, 0.5, 0.6},
			}
		} else {
			response = OllamaEmbedResponse{
				Embedding: []float64{0.0, 0.0, 0.0},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOllama,
		Model:          "nomic-embed-text",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     768,
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

	if response.Model != "nomic-embed-text" {
		t.Errorf("Expected model 'nomic-embed-text', got %s", response.Model)
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
	if len(embedding1.Embedding) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(embedding1.Embedding))
	}
	if embedding1.Embedding[0] != 0.1 || embedding1.Embedding[1] != 0.2 || embedding1.Embedding[2] != 0.3 {
		t.Errorf("Unexpected embedding values: %v", embedding1.Embedding)
	}

	// Verify second embedding
	embedding2 := response.Data[1]
	if embedding2.Index != 1 {
		t.Errorf("Expected index 1, got %d", embedding2.Index)
	}
	if len(embedding2.Embedding) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(embedding2.Embedding))
	}
	if embedding2.Embedding[0] != 0.4 || embedding2.Embedding[1] != 0.5 || embedding2.Embedding[2] != 0.6 {
		t.Errorf("Unexpected embedding values: %v", embedding2.Embedding)
	}

	// Verify usage
	if response.Usage.PromptTokens <= 0 {
		t.Errorf("Expected positive prompt tokens, got %d", response.Usage.PromptTokens)
	}
	if response.Usage.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens, got %d", response.Usage.TotalTokens)
	}
}

func TestOllamaProvider_CreateEmbeddingsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := OllamaEmbedResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOllama,
		Model:          "nomic-embed-text",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     768,
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
}

func TestOllamaProvider_CreateEmbeddingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOllama,
		Model:          "nomic-embed-text",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     0, // No retries for this test
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
		Dimensions:     768,
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

	if !strings.Contains(err.Error(), "failed to create embedding") {
		t.Errorf("Expected error to contain 'failed to create embedding', got: %v", err)
	}
}

func TestOllamaProvider_CreateEmbeddingsRetry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			// Fail first two requests
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on third request
		response := OllamaEmbedResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &embedding.Config{
		Provider:       embedding.ProviderOllama,
		Model:          "nomic-embed-text",
		BaseURL:        server.URL,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     10 * time.Millisecond, // Short delay for testing
		RateLimitRPS:   10.0,
		Dimensions:     768,
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

func TestOllamaProvider_EstimateTokens(t *testing.T) {
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

func TestOllamaProvider_SplitTextForEmbedding(t *testing.T) {
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

func TestOllamaProvider_ProviderInfo(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Test provider type
	if provider.GetProviderType() != embedding.ProviderOllama {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), embedding.ProviderOllama)
	}

	// Test model name
	if provider.GetModel() != "nomic-embed-text" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "nomic-embed-text")
	}

	// Test dimensions
	if provider.GetDimensions() != 768 {
		t.Errorf("GetDimensions() = %v, want %v", provider.GetDimensions(), 768)
	}

	// Test max tokens
	if provider.GetMaxTokens() != 8192 {
		t.Errorf("GetMaxTokens() = %v, want %v", provider.GetMaxTokens(), 8192)
	}
}

func TestOllamaProvider_EmptyTextList(t *testing.T) {
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
