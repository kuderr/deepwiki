package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kuderr/deepwiki/internal/logging"
	"github.com/kuderr/deepwiki/pkg/embedding"
	"golang.org/x/time/rate"
)

// OpenAIProvider implements embedding.Provider for OpenAI
type OpenAIProvider struct {
	config      *embedding.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger
}

// OpenAI API types
type EmbeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
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

// NewProvider creates a new OpenAI embedding provider
func NewProvider(config *embedding.Config) (embedding.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.Model == "" {
		return nil, fmt.Errorf("model is required")
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
		logger:      logging.GetGlobalLogger().WithComponent("openai-embed"),
	}

	return provider, nil
}

// CreateEmbeddings creates embeddings for the given texts
func (p *OpenAIProvider) CreateEmbeddings(
	ctx context.Context,
	texts []string,
	opts ...embedding.EmbeddingOptions,
) (*embedding.EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Apply options
	options := embedding.EmbeddingOptions{
		BatchSize: 100, // OpenAI supports up to 2048 texts per request
	}
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		options.BatchSize = opts[0].BatchSize
	}

	// If we have too many texts, batch them
	if len(texts) > options.BatchSize {
		return p.createEmbeddingsBatched(ctx, texts, options)
	}

	// Wait for rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	request := EmbeddingRequest{
		Model: p.config.Model,
		Input: texts,
	}

	// Set dimensions if specified in config
	if p.config.Dimensions > 0 {
		request.Dimensions = &p.config.Dimensions
	}

	p.logger.Debug("creating embeddings",
		slog.String("model", request.Model),
		slog.Int("text_count", len(texts)))

	response, err := p.sendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Convert to common format
	commonResponse := p.convertResponse(response)

	p.logger.Debug("embeddings created successfully",
		slog.Int("embedding_count", len(response.Data)),
		slog.Int("total_tokens", response.Usage.TotalTokens))

	return commonResponse, nil
}

// createEmbeddingsBatched handles large embedding requests by batching them
func (p *OpenAIProvider) createEmbeddingsBatched(
	ctx context.Context,
	texts []string,
	options embedding.EmbeddingOptions,
) (*embedding.EmbeddingResponse, error) {
	var allEmbeddings []embedding.Embedding
	var totalUsage embedding.Usage

	for i := 0; i < len(texts); i += options.BatchSize {
		end := i + options.BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		response, err := p.CreateEmbeddings(ctx, batch, options)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}

		// Adjust indices for the overall response
		for j, embeddingData := range response.Data {
			embeddingData.Index = i + j
			allEmbeddings = append(allEmbeddings, embeddingData)
		}

		totalUsage.PromptTokens += response.Usage.PromptTokens
		totalUsage.TotalTokens += response.Usage.TotalTokens
	}

	return &embedding.EmbeddingResponse{
		Object: "list",
		Data:   allEmbeddings,
		Model:  p.config.Model,
		Usage:  totalUsage,
	}, nil
}

// GetProviderType returns the provider type
func (p *OpenAIProvider) GetProviderType() embedding.ProviderType {
	return embedding.ProviderOpenAI
}

// GetModel returns the current model
func (p *OpenAIProvider) GetModel() string {
	return p.config.Model
}

// GetDimensions returns the embedding dimensions
func (p *OpenAIProvider) GetDimensions() int {
	if p.config.Dimensions > 0 {
		return p.config.Dimensions
	}

	// Default dimensions for OpenAI models
	switch p.config.Model {
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002":
		return 1536
	default:
		return 1536 // Default fallback
	}
}

// GetMaxTokens returns the maximum tokens supported
func (p *OpenAIProvider) GetMaxTokens() int {
	// OpenAI embedding models support up to 8191 tokens
	return 8191
}

// EstimateTokens estimates the number of tokens in the given text
func (p *OpenAIProvider) EstimateTokens(text string) int {
	// Simple approximation: ~4 characters per token
	// For precise counting, would need to use tiktoken
	return len(text) / 4
}

// SplitTextForEmbedding splits text into chunks that fit within token limits
func (p *OpenAIProvider) SplitTextForEmbedding(text string, maxTokens int) []string {
	if maxTokens <= 0 {
		maxTokens = p.GetMaxTokens()
	}

	// Handle empty text
	if text == "" {
		return []string{}
	}

	var chunks []string
	words := strings.Fields(text)
	var currentChunk strings.Builder

	for _, word := range words {
		// Check if adding this word would exceed the token limit
		estimatedTokens := p.EstimateTokens(currentChunk.String() + " " + word)
		if currentChunk.Len() > 0 && estimatedTokens > maxTokens {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(word)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// Helper methods

func (p *OpenAIProvider) sendRequest(ctx context.Context, request EmbeddingRequest) (*EmbeddingResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/embeddings"
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

	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &embeddingResponse, nil
}

func (p *OpenAIProvider) convertResponse(resp *EmbeddingResponse) *embedding.EmbeddingResponse {
	embeddings := make([]embedding.Embedding, len(resp.Data))
	for i, emb := range resp.Data {
		embeddings[i] = embedding.Embedding{
			Object:    emb.Object,
			Index:     emb.Index,
			Embedding: emb.Embedding,
		}
	}

	return &embedding.EmbeddingResponse{
		Object: resp.Object,
		Data:   embeddings,
		Model:  resp.Model,
		Usage: embedding.Usage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}
}
