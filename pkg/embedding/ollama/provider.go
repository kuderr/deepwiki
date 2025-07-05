package ollama

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

// OllamaProvider implements embedding.Provider for Ollama
type OllamaProvider struct {
	config      *embedding.Config
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *logging.Logger
}

// OllamaEmbedRequest represents a request to Ollama embeddings API
type OllamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbedResponse represents a response from Ollama embeddings API
type OllamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewProvider creates a new Ollama embedding provider
func NewProvider(config *embedding.Config) (embedding.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required for Ollama provider")
	}

	if config.Model == "" {
		return nil, fmt.Errorf("model is required")
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
		logger:      logging.GetGlobalLogger().WithComponent("ollama-embed"),
	}

	return provider, nil
}

// CreateEmbeddings creates embeddings for the given texts
func (p *OllamaProvider) CreateEmbeddings(
	ctx context.Context,
	texts []string,
	opts ...embedding.EmbeddingOptions,
) (*embedding.EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Apply options
	options := embedding.EmbeddingOptions{
		BatchSize: 10, // Ollama typically handles smaller batches better
	}
	if len(opts) > 0 && opts[0].BatchSize > 0 {
		options.BatchSize = opts[0].BatchSize
	}

	p.logger.Debug("creating embeddings",
		slog.String("model", p.config.Model),
		slog.Int("text_count", len(texts)),
		slog.Int("batch_size", options.BatchSize))

	var allEmbeddings []embedding.Embedding
	totalTokens := 0

	// Process texts in batches (Ollama processes one at a time)
	for i, text := range texts {
		if err := p.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiting failed: %w", err)
		}

		embeddingVec, err := p.createSingleEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}

		allEmbeddings = append(allEmbeddings, embedding.Embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: embeddingVec,
		})

		// Estimate tokens (roughly 4 characters per token)
		totalTokens += len(text) / 4
	}

	response := &embedding.EmbeddingResponse{
		Object: "list",
		Data:   allEmbeddings,
		Model:  p.config.Model,
		Usage: embedding.Usage{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	}

	p.logger.Debug("embeddings created successfully",
		slog.Int("embedding_count", len(allEmbeddings)),
		slog.Int("total_tokens", totalTokens))

	return response, nil
}

// createSingleEmbedding creates an embedding for a single text
func (p *OllamaProvider) createSingleEmbedding(ctx context.Context, text string) ([]float64, error) {
	request := OllamaEmbedRequest{
		Model:  p.config.Model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.config.BaseURL, "/") + "/api/embeddings"
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

		// Only log if we have an error and will retry
		if lastErr != nil && attempt < p.config.MaxRetries {
			p.logger.Warn("request attempt failed, retrying",
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", p.config.MaxRetries),
				slog.String("error", lastErr.Error()))
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
		return nil, fmt.Errorf("API request failed with status %d: %s", response.StatusCode, string(body))
	}

	var ollamaResponse OllamaEmbedResponse
	if err := json.Unmarshal(body, &ollamaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResponse.Embedding, nil
}

// GetProviderType returns the provider type
func (p *OllamaProvider) GetProviderType() embedding.ProviderType {
	return embedding.ProviderOllama
}

// GetModel returns the current model
func (p *OllamaProvider) GetModel() string {
	return p.config.Model
}

// GetDimensions returns the embedding dimensions
func (p *OllamaProvider) GetDimensions() int {
	if p.config.Dimensions > 0 {
		return p.config.Dimensions
	}
	// Default dimensions for common Ollama embedding models
	switch p.config.Model {
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	case "all-minilm":
		return 384
	default:
		return 768 // Default fallback
	}
}

// GetMaxTokens returns the maximum tokens supported
func (p *OllamaProvider) GetMaxTokens() int {
	// Most Ollama embedding models support around 512 tokens
	// but this varies by model
	switch p.config.Model {
	case "nomic-embed-text":
		return 8192
	case "mxbai-embed-large":
		return 512
	case "all-minilm":
		return 256
	default:
		return 512 // Conservative default
	}
}

// EstimateTokens estimates the number of tokens in the given text
func (p *OllamaProvider) EstimateTokens(text string) int {
	// Simple approximation: ~4 characters per token
	// This is a rough estimate - for precise counting, would need the actual tokenizer
	return len(text) / 4
}

// SplitTextForEmbedding splits text into chunks that fit within token limits
func (p *OllamaProvider) SplitTextForEmbedding(text string, maxTokens int) []string {
	if maxTokens <= 0 {
		maxTokens = p.GetMaxTokens()
	}

	// Handle empty text
	if text == "" {
		return []string{}
	}

	// Simple character-based splitting (approximating tokens)
	maxChars := maxTokens * 4 // ~4 chars per token

	if len(text) <= maxChars {
		return []string{text}
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
