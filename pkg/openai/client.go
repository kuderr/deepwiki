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
	"golang.org/x/time/rate"
)

// TODO: move to config
const (
	defaultBaseURL = "https://api.openai.com/v1"
)

// OpenAIClient implements the Client interface
type OpenAIClient struct {
	config      *Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger

	// Usage tracking
	usageMutex sync.RWMutex
	totalUsage TokenCount
}

// NewClient creates a new OpenAI client
func NewClient(config *Config) (*OpenAIClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Set up rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimitRPS), 1)

	// Set up HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	client := &OpenAIClient{
		config:      config,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		logger:      logging.GetGlobalLogger().WithComponent("openai"),
		totalUsage:  TokenCount{},
	}

	return client, nil
}

// ChatCompletion sends a chat completion request to OpenAI
func (c *OpenAIClient) ChatCompletion(
	ctx context.Context,
	messages []Message,
	opts ...ChatCompletionOptions,
) (*ChatCompletionResponse, error) {
	// Apply options
	options := ChatCompletionOptions{
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
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
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	request := ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      options.Stream,
	}

	c.logger.InfoContext(ctx, "sending chat completion request",
		slog.String("model", request.Model),
		slog.Int("message_count", len(messages)),
		slog.Int("max_tokens", request.MaxTokens),
		slog.Float64("temperature", request.Temperature),
		slog.Bool("stream", request.Stream),
	)

	// If streaming is requested, handle it differently
	if options.Stream && options.OnStream != nil {
		return nil, c.ChatCompletionStream(ctx, messages, options.OnStream, options)
	}

	response, err := c.doRequest(ctx, "POST", "/chat/completions", request)
	if err != nil {
		c.logger.LogError(ctx, "chat completion request failed", err,
			slog.String("model", request.Model),
		)
		return nil, err
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(response, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Track usage
	c.trackUsage(chatResponse.Usage)

	c.logger.InfoContext(ctx, "chat completion successful",
		slog.String("response_id", chatResponse.ID),
		slog.Int("prompt_tokens", chatResponse.Usage.PromptTokens),
		slog.Int("completion_tokens", chatResponse.Usage.CompletionTokens),
		slog.Int("total_tokens", chatResponse.Usage.TotalTokens),
	)

	return &chatResponse, nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *OpenAIClient) ChatCompletionStream(
	ctx context.Context,
	messages []Message,
	handler StreamHandler,
	opts ...ChatCompletionOptions,
) error {
	// Apply options
	options := ChatCompletionOptions{
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
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
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiting failed: %w", err)
	}

	request := ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		Stream:      true,
	}

	c.logger.InfoContext(ctx, "starting streaming chat completion",
		slog.String("model", request.Model),
		slog.Int("message_count", len(messages)),
	)

	return c.doStreamRequest(ctx, "POST", "/chat/completions", request, handler)
}

// CreateEmbeddings creates embeddings for the given texts
func (c *OpenAIClient) CreateEmbeddings(
	ctx context.Context,
	texts []string,
	opts ...EmbeddingOptions,
) (*EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Apply options
	options := EmbeddingOptions{
		BatchSize: 100, // Default batch size
	}
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		options.BatchSize = opts[0].BatchSize
	}

	// If we have too many texts, batch them
	if len(texts) > options.BatchSize {
		return c.createEmbeddingsBatched(ctx, texts, options)
	}

	// Wait for rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	request := EmbeddingRequest{
		Model: c.config.EmbeddingModel,
		Input: texts,
	}

	c.logger.InfoContext(ctx, "creating embeddings",
		slog.String("model", request.Model),
		slog.Int("text_count", len(texts)),
	)

	response, err := c.doRequest(ctx, "POST", "/embeddings", request)
	if err != nil {
		c.logger.LogError(ctx, "embedding request failed", err,
			slog.String("model", request.Model),
			slog.Int("text_count", len(texts)),
		)
		return nil, err
	}

	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(response, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Track usage
	c.trackUsage(embeddingResponse.Usage)

	c.logger.InfoContext(ctx, "embeddings created successfully",
		slog.Int("embedding_count", len(embeddingResponse.Data)),
		slog.Int("total_tokens", embeddingResponse.Usage.TotalTokens),
	)

	return &embeddingResponse, nil
}

// createEmbeddingsBatched handles large embedding requests by batching them
func (c *OpenAIClient) createEmbeddingsBatched(
	ctx context.Context,
	texts []string,
	options EmbeddingOptions,
) (*EmbeddingResponse, error) {
	var allEmbeddings []Embedding
	var totalUsage Usage

	for i := 0; i < len(texts); i += options.BatchSize {
		end := i + options.BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		response, err := c.CreateEmbeddings(ctx, batch, options)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}

		// Adjust indices for the overall response
		for j, embedding := range response.Data {
			embedding.Index = i + j
			allEmbeddings = append(allEmbeddings, embedding)
		}

		totalUsage.PromptTokens += response.Usage.PromptTokens
		totalUsage.CompletionTokens += response.Usage.CompletionTokens
		totalUsage.TotalTokens += response.Usage.TotalTokens
	}

	return &EmbeddingResponse{
		Object: "list",
		Data:   allEmbeddings,
		Model:  c.config.EmbeddingModel,
		Usage:  totalUsage,
	}, nil
}

// doRequest performs an HTTP request to the OpenAI API
func (c *OpenAIClient) doRequest(
	ctx context.Context,
	method, endpoint string,
	requestBody interface{},
) ([]byte, error) {
	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	url := defaultBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "deepwiki-cli/1.0")

	// Perform request with retries
	var response *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}

		response, lastErr = c.httpClient.Do(req)
		if lastErr == nil && response.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if response != nil {
			response.Body.Close()
		}

		c.logger.WarnContext(ctx, "request attempt failed, retrying",
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", c.config.MaxRetries),
			slog.String("error", lastErr.Error()),
		)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if response.StatusCode >= 400 {
		var apiError APIError
		if err := json.Unmarshal(responseBody, &apiError); err == nil {
			return nil, apiError
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

// doStreamRequest performs a streaming HTTP request
func (c *OpenAIClient) doStreamRequest(
	ctx context.Context,
	method, endpoint string,
	requestBody interface{},
	handler StreamHandler,
) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := defaultBaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for streaming
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("User-Agent", "deepwiki-cli/1.0")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(response.Body)
		var apiError APIError
		if err := json.Unmarshal(responseBody, &apiError); err == nil {
			return apiError
		}
		return fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(responseBody))
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse Server-Sent Events format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				break
			}

			var streamResponse StreamResponse
			if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
				c.logger.WarnContext(ctx, "failed to parse stream response",
					slog.String("data", data),
					slog.String("error", err.Error()),
				)
				continue
			}

			// Call the handler
			if err := handler(streamResponse); err != nil {
				return fmt.Errorf("stream handler error: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// trackUsage updates the total usage statistics
func (c *OpenAIClient) trackUsage(usage Usage) {
	c.usageMutex.Lock()
	defer c.usageMutex.Unlock()

	c.totalUsage.PromptTokens += usage.PromptTokens
	c.totalUsage.CompletionTokens += usage.CompletionTokens
	c.totalUsage.TotalTokens += usage.TotalTokens
	c.totalUsage.EstimatedCost += c.EstimateCost(usage.PromptTokens, usage.CompletionTokens)
}

// GetUsageStats returns the current usage statistics
func (c *OpenAIClient) GetUsageStats() TokenCount {
	c.usageMutex.RLock()
	defer c.usageMutex.RUnlock()
	return c.totalUsage
}

// ResetUsageStats resets the usage statistics
func (c *OpenAIClient) ResetUsageStats() {
	c.usageMutex.Lock()
	defer c.usageMutex.Unlock()
	c.totalUsage = TokenCount{}
}
