package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kuderr/deepwiki/internal/logging"
	"github.com/kuderr/deepwiki/pkg/llm"
	"golang.org/x/time/rate"
)

// OllamaProvider implements llm.Provider for Ollama
type OllamaProvider struct {
	config      *llm.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger

	// Usage tracking
	usageMutex sync.RWMutex
	totalUsage llm.TokenCount
}

// Ollama API types
type ChatCompletionRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

type StreamResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type APIError struct {
	Message string
}

func (e APIError) Error() string {
	return e.Message
}

// NewProvider creates a new Ollama LLM provider
func NewProvider(config *llm.Config) (llm.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}

	// Set up rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimitRPS), 1)

	// Set up HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	provider := &OllamaProvider{
		config:      config,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		logger:      logging.GetGlobalLogger().WithComponent("ollama-llm"),
		totalUsage:  llm.TokenCount{},
	}

	return provider, nil
}

// ChatCompletion sends a chat completion request to Ollama
func (p *OllamaProvider) ChatCompletion(
	ctx context.Context,
	messages []llm.Message,
	opts ...llm.ChatCompletionOptions,
) (*llm.ChatCompletionResponse, error) {
	// Apply options
	options := llm.ChatCompletionOptions{
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      false,
	}
	if len(opts) > 0 {
		if opts[0].MaxTokens > 0 {
			options.MaxTokens = opts[0].MaxTokens
		}
		if opts[0].Temperature >= 0 {
			options.Temperature = opts[0].Temperature
		}
		options.Stream = opts[0].Stream
	}

	// Wait for rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	// Convert messages
	ollamaMessages := make([]Message, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build options for Ollama
	requestOptions := make(map[string]interface{})
	if options.MaxTokens > 0 {
		requestOptions["num_predict"] = options.MaxTokens
	}
	if options.Temperature >= 0 {
		requestOptions["temperature"] = options.Temperature
	}

	request := ChatCompletionRequest{
		Model:    p.config.Model,
		Messages: ollamaMessages,
		Stream:   options.Stream,
		Options:  requestOptions,
	}

	p.logger.Debug("sending chat completion request",
		slog.String("model", request.Model),
		slog.Int("message_count", len(messages)),
		slog.Int("max_tokens", options.MaxTokens),
		slog.Float64("temperature", options.Temperature))

	response, err := p.sendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Convert response to common format
	commonResponse := p.convertToCommonResponse(response)

	// Update usage statistics
	p.updateUsageStats(response.PromptEvalCount, response.EvalCount)

	p.logger.Debug("chat completion successful",
		slog.Int("prompt_tokens", response.PromptEvalCount),
		slog.Int("completion_tokens", response.EvalCount))

	return commonResponse, nil
}

// ChatCompletionStream sends a streaming chat completion request
func (p *OllamaProvider) ChatCompletionStream(
	ctx context.Context,
	messages []llm.Message,
	handler llm.StreamHandler,
	opts ...llm.ChatCompletionOptions,
) error {
	// Apply options
	options := llm.ChatCompletionOptions{
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Stream:      true,
	}
	if len(opts) > 0 {
		if opts[0].MaxTokens > 0 {
			options.MaxTokens = opts[0].MaxTokens
		}
		if opts[0].Temperature >= 0 {
			options.Temperature = opts[0].Temperature
		}
	}

	// Wait for rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiting failed: %w", err)
	}

	// Convert messages
	ollamaMessages := make([]Message, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build options for Ollama
	requestOptions := make(map[string]interface{})
	if options.MaxTokens > 0 {
		requestOptions["num_predict"] = options.MaxTokens
	}
	if options.Temperature >= 0 {
		requestOptions["temperature"] = options.Temperature
	}

	request := ChatCompletionRequest{
		Model:    p.config.Model,
		Messages: ollamaMessages,
		Stream:   true,
		Options:  requestOptions,
	}

	return p.sendStreamingRequest(ctx, request, handler)
}

// CountTokens estimates token count for given text
func (p *OllamaProvider) CountTokens(text string) (int, error) {
	// Simple approximation: ~4 characters per token
	// Ollama doesn't provide a direct token counting API
	return len(text) / 4, nil
}

// EstimateCost estimates the cost (Ollama is typically free for local usage)
func (p *OllamaProvider) EstimateCost(promptTokens, completionTokens int) float64 {
	// Ollama is typically free for local usage
	return 0.0
}

// GetUsageStats returns current usage statistics
func (p *OllamaProvider) GetUsageStats() llm.TokenCount {
	p.usageMutex.RLock()
	defer p.usageMutex.RUnlock()
	return p.totalUsage
}

// ResetUsageStats resets usage statistics
func (p *OllamaProvider) ResetUsageStats() {
	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()
	p.totalUsage = llm.TokenCount{}
}

// GetProviderType returns the provider type
func (p *OllamaProvider) GetProviderType() llm.ProviderType {
	return llm.ProviderOllama
}

// GetModel returns the current model
func (p *OllamaProvider) GetModel() string {
	return p.config.Model
}

// Helper methods

func (p *OllamaProvider) sendRequest(
	ctx context.Context,
	request ChatCompletionRequest,
) (*ChatCompletionResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Perform request with retries
	var response *http.Response
	var lastErr error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.RetryDelay * time.Duration(attempt)):
			}
		}

		response, lastErr = p.httpClient.Do(req)
		if lastErr == nil && response.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if response != nil {
			response.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", p.config.MaxRetries, lastErr)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %s", errorResp.Error)
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResponse, nil
}

func (p *OllamaProvider) sendStreamingRequest(
	ctx context.Context,
	request ChatCompletionRequest,
	handler llm.StreamHandler,
) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("streaming request failed with status %d: %s", response.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			p.logger.Warn("failed to parse streaming response", slog.String("error", err.Error()))
			continue
		}

		// Convert to common format and call handler
		commonResp := p.convertStreamResponse(streamResp)
		if err := handler(commonResp); err != nil {
			return fmt.Errorf("stream handler error: %w", err)
		}

		// Break if done
		if streamResp.Done {
			break
		}
	}

	return scanner.Err()
}

func (p *OllamaProvider) convertToCommonResponse(resp *ChatCompletionResponse) *llm.ChatCompletionResponse {
	choices := []llm.Choice{
		{
			Index: 0,
			Message: llm.Message{
				Role:    resp.Message.Role,
				Content: resp.Message.Content,
			},
			FinishReason: "stop",
		},
	}

	// Convert timestamp to unix timestamp
	created := resp.CreatedAt.Unix()

	return &llm.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", created),
		Object:  "chat.completion",
		Created: created,
		Model:   resp.Model,
		Choices: choices,
		Usage: llm.Usage{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		},
	}
}

func (p *OllamaProvider) convertStreamResponse(resp StreamResponse) llm.StreamResponse {
	choices := []llm.StreamChoice{
		{
			Index: 0,
			Delta: llm.Message{
				Role:    resp.Message.Role,
				Content: resp.Message.Content,
			},
		},
	}

	// Convert timestamp to unix timestamp
	created := resp.CreatedAt.Unix()

	return llm.StreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", created),
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   resp.Model,
		Choices: choices,
	}
}

func (p *OllamaProvider) updateUsageStats(inputTokens, outputTokens int) {
	cost := p.EstimateCost(inputTokens, outputTokens)

	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()

	p.totalUsage.PromptTokens += inputTokens
	p.totalUsage.CompletionTokens += outputTokens
	p.totalUsage.TotalTokens += (inputTokens + outputTokens)
	p.totalUsage.EstimatedCost += cost
}
