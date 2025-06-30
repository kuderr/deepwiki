package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultConfig("test-api-key")
		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("NewClient failed: %v", err)
		}

		if client == nil {
			t.Fatal("Client is nil")
		}

		if client.config.APIKey != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', got '%s'", client.config.APIKey)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		_, err := NewClient(nil)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("empty API key", func(t *testing.T) {
		config := DefaultConfig("")
		_, err := NewClient(config)
		if err == nil {
			t.Error("Expected error for empty API key")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	apiKey := "test-key"
	config := DefaultConfig(apiKey)

	if config.APIKey != apiKey {
		t.Errorf("Expected API key '%s', got '%s'", apiKey, config.APIKey)
	}

	if config.Model == "" {
		t.Error("Expected default model to be set")
	}

	if config.EmbeddingModel == "" {
		t.Error("Expected default embedding model to be set")
	}

	if config.MaxTokens <= 0 {
		t.Error("Expected MaxTokens > 0")
	}

	if config.Temperature < 0 {
		t.Error("Expected Temperature >= 0")
	}

	if config.RequestTimeout <= 0 {
		t.Error("Expected RequestTimeout > 0")
	}

	if config.MaxRetries < 0 {
		t.Error("Expected MaxRetries >= 0")
	}

	if config.RetryDelay <= 0 {
		t.Error("Expected RetryDelay > 0")
	}
}

func TestChatCompletion(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path '/v1/chat/completions', got '%s'", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got '%s'", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Error("Missing or incorrect Authorization header")
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Missing or incorrect Content-Type header")
		}

		// Mock response
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 8,
				TotalTokens:      18,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom base URL
	config := DefaultConfig("test-api-key")
	config.RequestTimeout = 5 * time.Second

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the default base URL for testing
	originalHTTPClient := client.httpClient
	client.httpClient = &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &testTransport{
			baseURL: server.URL,
		},
	}
	defer func() { client.httpClient = originalHTTPClient }()

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	response, err := client.ChatCompletion(context.Background(), messages)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if len(response.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	if response.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected response content: %s", response.Choices[0].Message.Content)
	}

	// Check usage tracking
	stats := client.GetUsageStats()
	if stats.TotalTokens != 18 {
		t.Errorf("Expected total tokens 18, got %d", stats.TotalTokens)
	}
}

func TestCreateEmbeddings(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("Expected path '/v1/embeddings', got '%s'", r.URL.Path)
		}

		// Mock response
		response := EmbeddingResponse{
			Object: "list",
			Data: []Embedding{
				{
					Object:    "embedding",
					Index:     0,
					Embedding: make([]float64, 1536), // Standard embedding size
				},
				{
					Object:    "embedding",
					Index:     1,
					Embedding: make([]float64, 1536),
				},
			},
			Model: "text-embedding-3-small",
			Usage: Usage{
				PromptTokens: 15,
				TotalTokens:  15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := DefaultConfig("test-api-key")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the HTTP client for testing
	originalHTTPClient := client.httpClient
	client.httpClient = &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &testTransport{
			baseURL: server.URL,
		},
	}
	defer func() { client.httpClient = originalHTTPClient }()

	texts := []string{"Hello world", "Another text"}

	response, err := client.CreateEmbeddings(context.Background(), texts)
	if err != nil {
		t.Fatalf("CreateEmbeddings failed: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(response.Data))
	}

	if len(response.Data[0].Embedding) != 1536 {
		t.Errorf("Expected embedding dimension 1536, got %d", len(response.Data[0].Embedding))
	}
}

func TestTokenCounting(t *testing.T) {
	config := DefaultConfig("test-api-key")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		text          string
		expectedRange [2]int // min, max expected tokens
	}{
		{"", [2]int{0, 1}},
		{"Hello", [2]int{1, 5}},
		{"Hello world", [2]int{2, 8}},
		{"This is a longer text that should have more tokens", [2]int{10, 20}},
	}

	for _, test := range tests {
		tokens, err := client.CountTokens(test.text)
		if err != nil {
			t.Errorf("CountTokens failed for '%s': %v", test.text, err)
			continue
		}

		if tokens < test.expectedRange[0] || tokens > test.expectedRange[1] {
			t.Errorf("Token count for '%s' out of expected range [%d,%d], got %d",
				test.text, test.expectedRange[0], test.expectedRange[1], tokens)
		}
	}
}

func TestCostEstimation(t *testing.T) {
	config := DefaultConfig("test-api-key")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with known token counts
	promptTokens := 1000
	completionTokens := 500

	cost := client.EstimateCost(promptTokens, completionTokens)

	if cost <= 0 {
		t.Error("Expected cost > 0")
	}

	// Cost should be reasonable for these token counts (less than $1)
	if cost > 1.0 {
		t.Errorf("Cost seems too high: $%.4f for %d prompt + %d completion tokens",
			cost, promptTokens, completionTokens)
	}
}

func TestUsageTracking(t *testing.T) {
	config := DefaultConfig("test-api-key")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Initially, usage should be zero
	stats := client.GetUsageStats()
	if stats.TotalTokens != 0 {
		t.Errorf("Expected initial total tokens 0, got %d", stats.TotalTokens)
	}

	// Track some usage
	usage := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}
	client.trackUsage(usage)

	stats = client.GetUsageStats()
	if stats.TotalTokens != 150 {
		t.Errorf("Expected total tokens 150, got %d", stats.TotalTokens)
	}

	if stats.PromptTokens != 100 {
		t.Errorf("Expected prompt tokens 100, got %d", stats.PromptTokens)
	}

	if stats.CompletionTokens != 50 {
		t.Errorf("Expected completion tokens 50, got %d", stats.CompletionTokens)
	}

	// Reset and check
	client.ResetUsageStats()
	stats = client.GetUsageStats()
	if stats.TotalTokens != 0 {
		t.Errorf("Expected total tokens 0 after reset, got %d", stats.TotalTokens)
	}
}

func TestTruncateToTokenLimit(t *testing.T) {
	config := DefaultConfig("test-api-key")
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		text      string
		maxTokens int
		expectErr bool
	}{
		{
			name:      "short text within limit",
			text:      "Hello world",
			maxTokens: 100,
			expectErr: false,
		},
		{
			name:      "long text needs truncation",
			text:      strings.Repeat("This is a long text that needs to be truncated. ", 50),
			maxTokens: 50,
			expectErr: false,
		},
		{
			name:      "empty text",
			text:      "",
			maxTokens: 10,
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := client.TruncateToTokenLimit(test.text, test.maxTokens)

			if test.expectErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !test.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) > len(test.text) {
				t.Error("Truncated text is longer than original")
			}

			// Check that truncated text doesn't exceed token limit
			tokens, err := client.CountTokens(result)
			if err != nil {
				t.Errorf("Failed to count tokens in result: %v", err)
				return
			}

			if tokens > test.maxTokens {
				t.Errorf("Truncated text exceeds token limit: %d > %d", tokens, test.maxTokens)
			}
		})
	}
}

// testTransport is a custom RoundTripper for testing that redirects requests to a test server
type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL with our test server URL
	req.URL.Host = strings.TrimPrefix(t.baseURL, "http://")
	req.URL.Scheme = "http"

	return http.DefaultTransport.RoundTrip(req)
}
