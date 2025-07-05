package openai

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

// OpenAIProvider implements llm.Provider for OpenAI
type OpenAIProvider struct {
	config      *llm.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger

	// Usage tracking
	usageMutex sync.RWMutex
	totalUsage llm.TokenCount
}

// OpenAI API types
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Index int     `json:"index"`
	Delta Message `json:"delta"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type APIError struct {
	ErrorInfo struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (e APIError) Error() string {
	return e.ErrorInfo.Message
}

// NewProvider creates a new OpenAI LLM provider
func NewProvider(config *llm.Config) (llm.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	// Set up rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimitRPS), 1)

	// Set up HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	provider := &OpenAIProvider{
		config:      config,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		logger:      logging.GetGlobalLogger().WithComponent("openai-llm"),
		totalUsage:  llm.TokenCount{},
	}

	return provider, nil
}

// ChatCompletion sends a chat completion request to OpenAI
func (p *OpenAIProvider) ChatCompletion(
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
	openaiMessages := make([]Message, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	request := ChatCompletionRequest{
		Model:       p.config.Model,
		Messages:    openaiMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      options.Stream,
	}

	p.logger.Debug("sending chat completion request",
		slog.String("model", request.Model),
		slog.Int("message_count", len(messages)),
		slog.Int("max_tokens", request.MaxTokens),
		slog.Float64("temperature", request.Temperature))

	response, err := p.sendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Convert response to common format
	commonResponse := p.convertToCommonResponse(response)

	// Update usage statistics
	p.updateUsageStats(response.Usage.PromptTokens, response.Usage.CompletionTokens)

	p.logger.Debug("chat completion successful",
		slog.Int("prompt_tokens", response.Usage.PromptTokens),
		slog.Int("completion_tokens", response.Usage.CompletionTokens))

	return commonResponse, nil
}

// ChatCompletionStream sends a streaming chat completion request
func (p *OpenAIProvider) ChatCompletionStream(
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
	openaiMessages := make([]Message, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	request := ChatCompletionRequest{
		Model:       p.config.Model,
		Messages:    openaiMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      true,
	}

	return p.sendStreamingRequest(ctx, request, handler)
}

// CountTokens estimates token count for given text
func (p *OpenAIProvider) CountTokens(text string) (int, error) {
	// Simple approximation: ~4 characters per token
	// For precise counting, would need to use tiktoken
	return len(text) / 4, nil
}

// EstimateCost estimates the cost based on token usage
func (p *OpenAIProvider) EstimateCost(promptTokens, completionTokens int) float64 {
	// OpenAI pricing (these would need to be updated based on current pricing)
	var inputCostPer1M, outputCostPer1M float64

	switch p.config.Model {
	case "gpt-4o":
		inputCostPer1M = 2.50
		outputCostPer1M = 10.00
	case "gpt-4o-mini":
		inputCostPer1M = 0.15
		outputCostPer1M = 0.60
	case "gpt-3.5-turbo":
		inputCostPer1M = 0.50
		outputCostPer1M = 1.50
	default:
		// Default to gpt-4o pricing
		inputCostPer1M = 2.50
		outputCostPer1M = 10.00
	}

	inputCost := float64(promptTokens) * inputCostPer1M / 1000000
	outputCost := float64(completionTokens) * outputCostPer1M / 1000000
	return inputCost + outputCost
}

// GetUsageStats returns current usage statistics
func (p *OpenAIProvider) GetUsageStats() llm.TokenCount {
	p.usageMutex.RLock()
	defer p.usageMutex.RUnlock()
	return p.totalUsage
}

// ResetUsageStats resets usage statistics
func (p *OpenAIProvider) ResetUsageStats() {
	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()
	p.totalUsage = llm.TokenCount{}
}

// GetProviderType returns the provider type
func (p *OpenAIProvider) GetProviderType() llm.ProviderType {
	return llm.ProviderOpenAI
}

// GetModel returns the current model
func (p *OpenAIProvider) GetModel() string {
	return p.config.Model
}

// Helper methods

func (p *OpenAIProvider) sendRequest(
	ctx context.Context,
	request ChatCompletionRequest,
) (*ChatCompletionResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API error: %w", apiError)
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResponse, nil
}

func (p *OpenAIProvider) sendStreamingRequest(
	ctx context.Context,
	request ChatCompletionRequest,
	handler llm.StreamHandler,
) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

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

func (p *OpenAIProvider) convertToCommonResponse(resp *ChatCompletionResponse) *llm.ChatCompletionResponse {
	choices := make([]llm.Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = llm.Choice{
			Index: choice.Index,
			Message: llm.Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &llm.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

func (p *OpenAIProvider) convertStreamResponse(resp StreamResponse) llm.StreamResponse {
	choices := make([]llm.StreamChoice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = llm.StreamChoice{
			Index: choice.Index,
			Delta: llm.Message{
				Role:    choice.Delta.Role,
				Content: choice.Delta.Content,
			},
		}
	}

	return llm.StreamResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
	}
}

func (p *OpenAIProvider) updateUsageStats(inputTokens, outputTokens int) {
	cost := p.EstimateCost(inputTokens, outputTokens)

	p.usageMutex.Lock()
	defer p.usageMutex.Unlock()

	p.totalUsage.PromptTokens += inputTokens
	p.totalUsage.CompletionTokens += outputTokens
	p.totalUsage.TotalTokens += (inputTokens + outputTokens)
	p.totalUsage.EstimatedCost += cost
}
