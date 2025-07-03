package anthropic

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
				Provider: llm.ProviderAnthropic,
				Model:    "claude-3-5-sonnet-20241022",
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "valid config",
			config: &llm.Config{
				Provider:       llm.ProviderAnthropic,
				APIKey:         "test-key",
				Model:          "claude-3-5-sonnet-20241022",
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
				Provider:       llm.ProviderAnthropic,
				APIKey:         "test-key",
				Model:          "claude-3-5-sonnet-20241022",
				BaseURL:        "https://custom.anthropic.com/v1",
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
			if provider.GetProviderType() != llm.ProviderAnthropic {
				t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderAnthropic)
			}

			if provider.GetModel() != tt.config.Model {
				t.Errorf("GetModel() = %v, want %v", provider.GetModel(), tt.config.Model)
			}

			// Test default base URL is set
			if tt.config.BaseURL == "" &&
				provider.(*AnthropicProvider).config.BaseURL != "https://api.anthropic.com/v1" {
				t.Errorf("Default BaseURL not set correctly")
			}
		})
	}
}

func TestAnthropicProvider_ChatCompletion(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("Expected x-api-key header to be 'test-key', got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != anthropicVersion {
			t.Errorf(
				"Expected anthropic-version header to be %s, got %s",
				anthropicVersion,
				r.Header.Get("anthropic-version"),
			)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Mock response
		response := MessagesResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-3-5-sonnet-20241022",
			Content: []Content{
				{Type: "text", Text: "Hello! How can I help you today?"},
			},
			StopReason: "end_turn",
			Usage: Usage{
				InputTokens:  10,
				OutputTokens: 15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

	if response.ID != "msg_123" {
		t.Errorf("Expected response ID 'msg_123', got %s", response.ID)
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

func TestAnthropicProvider_ChatCompletionWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to verify options
		var request MessagesRequest
		json.NewDecoder(r.Body).Decode(&request)

		if request.MaxTokens != 1000 {
			t.Errorf("Expected max tokens 1000, got %d", request.MaxTokens)
		}
		if request.Temperature != 0.5 {
			t.Errorf("Expected temperature 0.5, got %f", request.Temperature)
		}

		response := MessagesResponse{
			ID:    "msg_123",
			Type:  "message",
			Role:  "assistant",
			Model: "claude-3-5-sonnet-20241022",
			Content: []Content{
				{Type: "text", Text: "Test response"},
			},
			StopReason: "end_turn",
			Usage: Usage{
				InputTokens:  5,
				OutputTokens: 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

func TestAnthropicProvider_ChatCompletionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := ErrorResponse{
			Error: APIError{
				Type:    "invalid_request_error",
				Message: "Invalid request",
			},
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

func TestAnthropicProvider_ChatCompletionStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify streaming request
		var request MessagesRequest
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
			`data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}`,
			`data: {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": " world"}}`,
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
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

func TestAnthropicProvider_CountTokens(t *testing.T) {
	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

func TestAnthropicProvider_EstimateCost(t *testing.T) {
	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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

	// Expected cost: (1000 * 3.0 / 1000000) + (500 * 15.0 / 1000000) = 0.003 + 0.0075 = 0.0105
	expectedCost := 0.0105
	tolerance := 0.000001
	if cost < expectedCost-tolerance || cost > expectedCost+tolerance {
		t.Errorf("EstimateCost() = %.6f, want %.6f (±%.6f)", cost, expectedCost, tolerance)
	}
}

func TestAnthropicProvider_UsageStats(t *testing.T) {
	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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
	anthropicProvider := provider.(*AnthropicProvider)
	anthropicProvider.updateUsageStats(100, 50)

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

func TestAnthropicProvider_ProviderInfo(t *testing.T) {
	config := &llm.Config{
		Provider:       llm.ProviderAnthropic,
		APIKey:         "test-key",
		Model:          "claude-3-5-sonnet-20241022",
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
	if provider.GetProviderType() != llm.ProviderAnthropic {
		t.Errorf("GetProviderType() = %v, want %v", provider.GetProviderType(), llm.ProviderAnthropic)
	}

	// Test model name
	if provider.GetModel() != "claude-3-5-sonnet-20241022" {
		t.Errorf("GetModel() = %v, want %v", provider.GetModel(), "claude-3-5-sonnet-20241022")
	}
}
