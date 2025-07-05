package anthropic

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

const anthropicVersion = "2023-06-01"

// AnthropicProvider implements llm.Provider for Anthropic Claude API
type AnthropicProvider struct {
	config      *llm.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger

	// Usage tracking
	usageMutex sync.RWMutex
	totalUsage llm.TokenCount
}

// Anthropic API request/response types
type MessagesRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
	System      string    `json:"system,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessagesResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	Content      []Content `json:"content"`
	StopReason   string    `json:"stop_reason"`
	StopSequence string    `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type StreamResponse struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta *Delta `json:"delta,omitempty"`
	Usage *Usage `json:"usage,omitempty"`
}

type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewProvider creates a new Anthropic LLM provider
func NewProvider(config *llm.Config) (llm.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}

	// Set up rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimitRPS), 1)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	provider := &AnthropicProvider{
		config:      config,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		logger:      logging.GetGlobalLogger().WithComponent("anthropic-llm"),
		totalUsage:  llm.TokenCount{},
	}

	return provider, nil
}

// ChatCompletion sends a chat completion request to Anthropic
func (p *AnthropicProvider) ChatCompletion(
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

	// Convert messages to Anthropic format
	anthropicMessages := make([]Message, len(messages))
	for i, msg := range messages {
		anthropicMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	request := MessagesRequest{
		Model:       p.config.Model,
		Messages:    anthropicMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      false,
	}

	response, err := p.sendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Update usage stats
	p.updateUsageStats(response.Usage.InputTokens, response.Usage.OutputTokens)

	p.logger.InfoContext(ctx, "chat completion successful",
		slog.Int("input_tokens", response.Usage.InputTokens),
		slog.Int("output_tokens", response.Usage.OutputTokens),
	)

	// Convert response to common format
	return p.convertResponse(response), nil
}

// ChatCompletionStream sends a streaming chat completion request
func (p *AnthropicProvider) ChatCompletionStream(
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

	// Convert messages to Anthropic format
	anthropicMessages := make([]Message, len(messages))
	for i, msg := range messages {
		anthropicMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	request := MessagesRequest{
		Model:       p.config.Model,
		Messages:    anthropicMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      true,
	}

	return p.sendStreamRequest(ctx, request, handler)
}

// CountTokens estimates the number of tokens in a text
func (p *AnthropicProvider) CountTokens(text string) (int, error) {
	// Anthropic uses a similar tokenization to GPT models
	// Rough estimation: ~4 characters per token
	return len(text) / 4, nil
}

// EstimateCost estimates the cost for the given token usage
func (p *AnthropicProvider) EstimateCost(promptTokens, completionTokens int) float64 {
	// Claude 3.5 Sonnet pricing (as of 2024)
	// Input: $3 per 1M tokens, Output: $15 per 1M tokens
	inputCost := float64(promptTokens) * 3.0 / 1000000
	outputCost := float64(completionTokens) * 15.0 / 1000000
	return inputCost + outputCost
}

// GetUsageStats returns current usage statistics
func (p *AnthropicProvider) GetUsageStats() llm.TokenCount {
	p.usageMutex.RLock()
	defer p.usageMutex.RUnlock()
	return p.totalUsage
}

// ResetUsageStats resets usage statistics
func (p *AnthropicProvider) ResetUsageStats() {
	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()
	p.totalUsage = llm.TokenCount{}
}

// GetProviderType returns the provider type
func (p *AnthropicProvider) GetProviderType() llm.ProviderType {
	return llm.ProviderAnthropic
}

// GetModel returns the model name
func (p *AnthropicProvider) GetModel() string {
	return p.config.Model
}

// Helper methods

func (p *AnthropicProvider) sendRequest(ctx context.Context, request MessagesRequest) (*MessagesResponse, error) {
	// Rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/messages", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	// Send request with retries
	var response *MessagesResponse
	var lastErr error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.config.RetryDelay * time.Duration(attempt)):
			}
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode >= 400 {
			var errorResp ErrorResponse
			if err := json.Unmarshal(body, &errorResp); err == nil {
				lastErr = fmt.Errorf("API error: %s", errorResp.Error.Message)
			} else {
				lastErr = fmt.Errorf("API error: %s", string(body))
			}
			continue
		}

		if err := json.Unmarshal(body, &response); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

func (p *AnthropicProvider) sendStreamRequest(
	ctx context.Context,
	request MessagesRequest,
	handler llm.StreamHandler,
) error {
	// Rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/messages", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return fmt.Errorf("API error: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("API error: %s", string(body))
	}

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			p.logger.Warn("failed to parse streaming response", slog.String("error", err.Error()))
			continue
		}

		// Convert to common format and call handler
		commonResp := p.convertStreamResponse(streamResp)
		if err := handler(commonResp); err != nil {
			return fmt.Errorf("stream handler error: %w", err)
		}
	}

	return scanner.Err()
}

func (p *AnthropicProvider) updateUsageStats(inputTokens, outputTokens int) {
	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()

	p.totalUsage.PromptTokens += inputTokens
	p.totalUsage.CompletionTokens += outputTokens
	p.totalUsage.TotalTokens += inputTokens + outputTokens
	p.totalUsage.EstimatedCost += p.EstimateCost(inputTokens, outputTokens)
}

func (p *AnthropicProvider) convertResponse(resp *MessagesResponse) *llm.ChatCompletionResponse {
	var content string
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}

	return &llm.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    resp.Role,
					Content: content,
				},
				FinishReason: resp.StopReason,
			},
		},
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

func (p *AnthropicProvider) convertStreamResponse(resp StreamResponse) llm.StreamResponse {
	var deltaText string
	if resp.Delta != nil {
		deltaText = resp.Delta.Text
	}

	return llm.StreamResponse{
		ID:      "stream-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   p.config.Model,
		Choices: []llm.StreamChoice{
			{
				Index: resp.Index,
				Delta: llm.Message{
					Role:    "assistant",
					Content: deltaText,
				},
			},
		},
	}
}
