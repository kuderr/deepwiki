package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/llm"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  *llm.Config
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
			config: &llm.Config{
				Provider: llm.ProviderOpenAI,
				Model:    "gpt-4o",
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "valid config",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "gpt-4o",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
			},
			wantErr: false,
		},
		{
			name: "config with custom base URL",
			config: &llm.Config{
				Provider:       llm.ProviderOpenAI,
				APIKey:         "test-key",
				Model:          "gpt-4o",
				BaseURL:        "https://custom.openai.com/v1",
				MaxTokens:      4000,
				Temperature:    0.1,
				RequestTimeout: 30 * time.Second,
				MaxRetries:     3,
				RetryDelay:     1 * time.Second,
				RateLimitRPS:   2.0,
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
			if provider.GetProviderType() != llm.ProviderOpenAI {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderOpenAI)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}

			// Test default base URL is set
			if tt.config.BaseURL == "" && provider.(*OpenAIProvider).config.BaseURL != "https://api.openai.com/v1" {
				t.Errorf("Default BaseURL not set correctly")
			}
		})
	}
}

func TestOpenAIProvider_ChatCompletion(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Mock response
		response := ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4o",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Hello! How can I help you today?",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 15,
				TotalTokens:      25,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "gpt-4o",
		BaseURL:        server.URL,
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0, // High rate limit for tests
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	messages := []llm.Message{
		{Role: "user", Content: "Hello"},
	}

	response, err := provider.ChatCompletion(ctx, messages)
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}

	// Verify response
	if response == nil {
		t.Fatal("ChatCompletion() returned nil response")
	}

	if response.ID != "chatcmpl-123" {
		t.Errorf("Expected response ID 'chatcmpl-123', got %s", response.ID)
	}

	if len(response.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(response.Choices))
	}

	choice := response.Choices[0]
	if choice.Message.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected message content 'Hello! How can I help you today?', got %s", choice.Message.Content)
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected message role 'assistant', got %s", choice.Message.Role)
	}

	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", response.Usage.PromptTokens)
	}

	if response.Usage.CompletionTokens != 15 {
		t.Errorf("Expected completion tokens 15, got %d", response.Usage.CompletionTokens)
	}

	if response.Usage.TotalTokens != 25 {
		t.Errorf("Expected total tokens 25, got %d", response.Usage.TotalTokens)
	}

	// Verify usage stats were updated
	stats := provider.GetUsageStats()
	if stats.PromptTokens != 10 {
		t.Errorf("Expected usage stats prompt tokens 10, got %d", stats.PromptTokens)
	}
	if stats.CompletionTokens != 15 {
		t.Errorf("Expected usage stats completion tokens 15, got %d", stats.CompletionTokens)
	}
	if stats.TotalTokens != 25 {
		t.Errorf("Expected usage stats total tokens 25, got %d", stats.TotalTokens)
	}
	if stats.EstimatedCost <= 0 {
		t.Errorf("Expected positive estimated cost, got %f", stats.EstimatedCost)
	}
}

func TestOpenAIProvider_ChatCompletionWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to verify options
		var request ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&request)

		if request.MaxTokens != 1000 {
			t.Errorf("Expected max tokens 1000, got %d", request.MaxTokens)
		}
		if request.Temperature != 0.5 {
			t.Errorf("Expected temperature 0.5, got %f", request.Temperature)
		}

		response := ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4o",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Test response",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     5,
				CompletionTokens: 10,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "gpt-4o",
		BaseURL:        server.URL,
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	messages := []llm.Message{
		{Role: "user", Content: "Test"},
	}

	options := llm.ChatCompletionOptions{
		MaxTokens:   1000,
		Temperature: 0.5,
	}

	response, err := provider.ChatCompletion(ctx, messages, options)
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}

	if response == nil {
		t.Fatal("ChatCompletion() returned nil response")
	}
}

func TestOpenAIProvider_ChatCompletionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: ErrorDetail{
				Message: "Invalid request",
				Type:    "invalid_request_error",
				Code:    "bad_request",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "gpt-4o",
		BaseURL:        server.URL,
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     0, // No retries for this test
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	messages := []llm.Message{
		{Role: "user", Content: "Test"},
	}

	_, err = provider.ChatCompletion(ctx, messages)
	if err == nil {
		t.Fatal("Expected error from ChatCompletion(), got nil")
	}

	if !strings.Contains(err.Error(), "Invalid request") {
		t.Errorf("Expected error to contain 'Invalid request', got: %v", err)
	}
}

func TestOpenAIProvider_ChatCompletionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify streaming request
		var request ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&request)

		if !request.Stream {
			t.Error("Expected stream to be true")
		}

		// Send streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send some streaming data
		streamResponses := []string{
			`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
			`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}`,
			`data: [DONE]`,
		}

		for _, response := range streamResponses {
			w.Write([]byte(response + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderOpenAI,
		APIKey:         "test-key",
		Model:          "gpt-4o",
		BaseURL:        server.URL,
		MaxTokens:      4000,
		Temperature:    0.1,
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10.0,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	ctx := context.Background()
	messages := []llm.Message{
		{Role: "user", Content: "Hello"},
	}

	var receivedChunks []llm.StreamResponse
	handler := func(chunk llm.StreamResponse) error {
		receivedChunks = append(receivedChunks, chunk)
		return nil
	}

	err = provider.ChatCompletionStream(ctx, messages, handler)
	if err != nil {
		t.Fatalf("ChatCompletionStream() error = %v", err)
	}

	if len(receivedChunks) != 2 {
		t.Errorf("Expected 2 stream chunks, got %d", len(receivedChunks))
	}

	// Verify first chunk
	if len(receivedChunks) > 0 {
		chunk := receivedChunks[0]
		if chunk.Object != "chat.completion.chunk" {
			t.Errorf("Expected object 'chat.completion.chunk', got %s", chunk.Object)
		}
		if len(chunk.Choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(chunk.Choices))
		}
		if chunk.Choices[0].Delta.Content != "Hello" {
			t.Errorf("Expected delta content 'Hello', got %s", chunk.Choices[0].Delta.Content)
		}
	}
}

func TestOpenAIProvider_CountTokens(t *testing.T) {
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
			count, err := provider.CountTokens(tt.text)
			if err != nil {
				t.Errorf("CountTokens() error = %v", err)
			}
			if count != tt.expected {
				t.Errorf("CountTokens() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestOpenAIProvider_EstimateCost(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Test cost estimation
	promptTokens := 1000
	completionTokens := 500

	cost := provider.EstimateCost(promptTokens, completionTokens)

	// Expected cost for gpt-4o: (1000 * 2.5 / 1000000) + (500 * 10.0 / 1000000) = 0.0025 + 0.005 = 0.0075
	expectedCost := 0.0075
	tolerance := 0.000001
	if cost < expectedCost-tolerance || cost > expectedCost+tolerance {
		t.Errorf("EstimateCost() = %.6f, want %.6f (±%.6f)", cost, expectedCost, tolerance)
	}
}

func TestOpenAIProvider_UsageStats(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Initial stats should be zero
	stats := provider.GetUsageStats()
	if stats.PromptTokens != 0 || stats.CompletionTokens != 0 || stats.TotalTokens != 0 || stats.EstimatedCost != 0 {
		t.Errorf("Initial usage stats should be zero, got %+v", stats)
	}

	// Simulate usage update
	openaiProvider := provider.(*OpenAIProvider)
	openaiProvider.updateUsageStats(100, 50)

	// Check updated stats
	stats = provider.GetUsageStats()
	if stats.PromptTokens != 100 {
		t.Errorf("Expected prompt tokens 100, got %d", stats.PromptTokens)
	}
	if stats.CompletionTokens != 50 {
		t.Errorf("Expected completion tokens 50, got %d", stats.CompletionTokens)
	}
	if stats.TotalTokens != 150 {
		t.Errorf("Expected total tokens 150, got %d", stats.TotalTokens)
	}
	if stats.EstimatedCost <= 0 {
		t.Errorf("Expected positive estimated cost, got %f", stats.EstimatedCost)
	}

	// Test reset
	provider.ResetUsageStats()
	stats = provider.GetUsageStats()
	if stats.PromptTokens != 0 || stats.CompletionTokens != 0 || stats.TotalTokens != 0 || stats.EstimatedCost != 0 {
		t.Errorf("Usage stats should be zero after reset, got %+v", stats)
	}
}

func TestOpenAIProvider_ProviderInfo(t *testing.T) {
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

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Test provider type
	if provider.GetProviderType() != llm.ProviderOpenAI {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderOpenAI)
	}

	// Test model name
	if provider.GetModel() != "gpt-4o" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "gpt-4o")
	}
}
