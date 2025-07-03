package voyage

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

// VoyageProvider implements embedding.Provider for Voyage AI
type VoyageProvider struct {
	config      *embedding.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger
}

// Voyage AI API types
type VoyageEmbeddingRequest struct {
	Input      []string `json:"input"`
	Model      string   `json:"model"`
	InputType  string   `json:"input_type,omitempty"`
	Truncation bool     `json:"truncation,omitempty"`
}

type VoyageEmbeddingResponse struct {
	Object string            `json:"object"`
	Data   []VoyageEmbedding `json:"data"`
	Model  string            `json:"model"`
	Usage  VoyageUsage       `json:"usage"`
}

type VoyageEmbedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type VoyageUsage struct {
	TotalTokens int `json:"total_tokens"`
}

// Type aliases for test compatibility
type (
	EmbeddingRequest  = VoyageEmbeddingRequest
	EmbeddingResponse = VoyageEmbeddingResponse
	EmbeddingData     = VoyageEmbedding
	EmbeddingUsage    = VoyageUsage
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Detail string `json:"detail"`
}

type APIError struct {
	Detail string `json:"detail"`
}

func (e APIError) Error() string {
	return e.Detail
}

// NewProvider creates a new Voyage AI embedding provider
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
		config.BaseURL = "https://api.voyageai.com/v1"
	}

	// Set up rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.RateLimitRPS), 1)

	// Set up HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
	}

	provider := &VoyageProvider{
		config:      config,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		logger:      logging.GetGlobalLogger().WithComponent("voyage-embed"),
	}

	return provider, nil
}

// CreateEmbeddings creates embeddings for the given texts
func (p *VoyageProvider) CreateEmbeddings(
	ctx context.Context,
	texts []string,
	opts ...embedding.EmbeddingOptions,
) (*embedding.EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Apply options
	options := embedding.EmbeddingOptions{
		BatchSize: 128,        // Voyage AI supports up to 128 texts per request
		InputType: "document", // Default to document type
	}
	if len(opts) > 0 {
		if opts[0].BatchSize > 0 {
			options.BatchSize = opts[0].BatchSize
		}
		if opts[0].InputType != "" {
			options.InputType = opts[0].InputType
		}
	}

	// If we have too many texts, batch them
	if len(texts) > options.BatchSize {
		return p.createEmbeddingsBatched(ctx, texts, options)
	}

	// Wait for rate limiting
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiting failed: %w", err)
	}

	request := VoyageEmbeddingRequest{
		Input:     texts,
		Model:     p.config.Model,
		InputType: options.InputType,
	}

	p.logger.Debug("creating embeddings",
		slog.String("model", request.Model),
		slog.Int("text_count", len(texts)),
		slog.String("input_type", options.InputType))

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
func (p *VoyageProvider) createEmbeddingsBatched(
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
func (p *VoyageProvider) GetProviderType() embedding.ProviderType {
	return embedding.ProviderVoyage
}

// GetModel returns the current model
func (p *VoyageProvider) GetModel() string {
	return p.config.Model
}

// GetDimensions returns the embedding dimensions
func (p *VoyageProvider) GetDimensions() int {
	if p.config.Dimensions > 0 {
		return p.config.Dimensions
	}

	// Default dimensions for Voyage models
	switch p.config.Model {
	case "voyage-3-large":
		return 1024
	case "voyage-3.5-lite":
		return 512
	case "voyage-3.5":
		return 1024
	case "voyage-code-3":
		return 1024
	case "voyage-finance-2":
		return 1024
	default:
		return 1024 // Default fallback
	}
}

// GetMaxTokens returns the maximum tokens supported
func (p *VoyageProvider) GetMaxTokens() int {
	// Different Voyage models have different token limits
	switch p.config.Model {
	case "voyage-code-3":
		return 16000
	default:
		return 32000 // Default for most Voyage models
	}
}

// EstimateTokens estimates the number of tokens in the given text
func (p *VoyageProvider) EstimateTokens(text string) int {
	// Simple approximation: ~4 characters per token
	return len(text) / 4
}

// SplitTextForEmbedding splits text into chunks that fit within token limits
func (p *VoyageProvider) SplitTextForEmbedding(text string, maxTokens int) []string {
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

func (p *VoyageProvider) sendRequest(
	ctx context.Context,
	request VoyageEmbeddingRequest,
) (*VoyageEmbeddingResponse, error) {
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
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			// Try fallback to simple APIError
			var apiError APIError
			if err := json.Unmarshal(body, &apiError); err != nil {
				return nil, fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
			}
			return nil, fmt.Errorf("API error: %w", apiError)
		}
		return nil, fmt.Errorf("API error: %s", errorResp.Error.Detail)
	}

	var embeddingResponse VoyageEmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &embeddingResponse, nil
}

func (p *VoyageProvider) convertResponse(resp *VoyageEmbeddingResponse) *embedding.EmbeddingResponse {
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
			PromptTokens: resp.Usage.TotalTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}
}
