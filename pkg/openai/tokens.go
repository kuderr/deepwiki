package openai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// Pricing constants (as of June 2025, in USD per 1K tokens)
const (
	// GPT-4o pricing
	GPT4OInputPrice  = 0.005 // $5.00 per 1M tokens
	GPT4OOutputPrice = 0.015 // $15.00 per 1M tokens

	// GPT-4 Turbo pricing
	GPT4TurboInputPrice  = 0.01 // $10.00 per 1M tokens
	GPT4TurboOutputPrice = 0.03 // $30.00 per 1M tokens

	// GPT-3.5 Turbo pricing
	GPT35TurboInputPrice  = 0.0005 // $0.50 per 1M tokens
	GPT35TurboOutputPrice = 0.0015 // $1.50 per 1M tokens

	// Embedding pricing
	TextEmbedding3SmallPrice = 0.00002 // $0.02 per 1M tokens
	TextEmbedding3LargePrice = 0.00013 // $0.13 per 1M tokens
	AdaEmbeddingPrice        = 0.0001  // $0.10 per 1M tokens
)

// CountTokens provides accurate token count using tiktoken
func (c *OpenAIClient) CountTokens(text string) (int, error) {
	if text == "" {
		return 0, nil
	}

	// Get the appropriate encoding for the model
	encoding, err := c.getTokenEncoding()
	if err != nil {
		// Fallback to estimation if tiktoken fails
		c.logger.WarnContext(context.Background(), "failed to get tiktoken encoding, using estimation",
			slog.String("model", c.config.Model),
			slog.String("error", err.Error()),
		)
		return estimateTokenCount(text), nil
	}

	tokens := encoding.EncodeOrdinary(text)
	return len(tokens), nil
}

// tokenEncodingCache caches tiktoken encodings for performance
var (
	tokenEncodingCache = make(map[string]*tiktoken.Tiktoken)
	tokenEncodingMutex sync.RWMutex
)

// getTokenEncoding returns the appropriate tiktoken encoding for the current model
func (c *OpenAIClient) getTokenEncoding() (*tiktoken.Tiktoken, error) {
	model := c.config.Model

	tokenEncodingMutex.RLock()
	if encoding, exists := tokenEncodingCache[model]; exists {
		tokenEncodingMutex.RUnlock()
		return encoding, nil
	}
	tokenEncodingMutex.RUnlock()

	tokenEncodingMutex.Lock()
	defer tokenEncodingMutex.Unlock()

	// Double-check in case another goroutine added it
	if encoding, exists := tokenEncodingCache[model]; exists {
		return encoding, nil
	}

	var encoding *tiktoken.Tiktoken
	var err error

	// Try to get encoding by model name first
	encoding, err = tiktoken.EncodingForModel(model)
	if err != nil {
		// Fallback to appropriate encoding based on model family
		switch {
		case strings.Contains(strings.ToLower(model), "gpt-4"):
			encoding, err = tiktoken.GetEncoding("cl100k_base")
		case strings.Contains(strings.ToLower(model), "gpt-3.5"):
			encoding, err = tiktoken.GetEncoding("cl100k_base")
		default:
			// Default to cl100k_base for modern models
			encoding, err = tiktoken.GetEncoding("cl100k_base")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get encoding for model %s: %w", model, err)
	}

	tokenEncodingCache[model] = encoding
	return encoding, nil
}

// EstimateCost calculates the estimated cost for token usage
func (c *OpenAIClient) EstimateCost(promptTokens, completionTokens int) float64 {
	inputPrice, outputPrice := c.getModelPricing()

	inputCost := float64(promptTokens) * inputPrice / 1000.0
	outputCost := float64(completionTokens) * outputPrice / 1000.0

	return inputCost + outputCost
}

// getModelPricing returns input and output pricing for the current model
func (c *OpenAIClient) getModelPricing() (inputPrice, outputPrice float64) {
	model := strings.ToLower(c.config.Model)

	switch {
	case strings.Contains(model, "gpt-4o"):
		return GPT4OInputPrice, GPT4OOutputPrice
	case strings.Contains(model, "gpt-4-turbo") || strings.Contains(model, "gpt-4-1106") || strings.Contains(model, "gpt-4-0125"):
		return GPT4TurboInputPrice, GPT4TurboOutputPrice
	case strings.Contains(model, "gpt-4"):
		return GPT4TurboInputPrice, GPT4TurboOutputPrice // Use turbo pricing as fallback
	case strings.Contains(model, "gpt-3.5-turbo"):
		return GPT35TurboInputPrice, GPT35TurboOutputPrice
	default:
		// Default to GPT-4o pricing
		c.logger.WarnContext(context.Background(), "unknown model, using GPT-4o pricing",
			slog.String("model", c.config.Model),
		)
		return GPT4OInputPrice, GPT4OOutputPrice
	}
}

// GetEmbeddingPricing returns the pricing for embedding models
func (c *OpenAIClient) GetEmbeddingPricing() float64 {
	model := strings.ToLower(c.config.EmbeddingModel)

	switch {
	case strings.Contains(model, "text-embedding-3-small"):
		return TextEmbedding3SmallPrice
	case strings.Contains(model, "text-embedding-3-large"):
		return TextEmbedding3LargePrice
	case strings.Contains(model, "text-embedding-ada-002"):
		return AdaEmbeddingPrice
	default:
		return TextEmbedding3SmallPrice // Default to most common model
	}
}

// estimateTokenCount provides a rough estimate of token count
// This is a temporary implementation until tiktoken-go is integrated
func estimateTokenCount(text string) int {
	if text == "" {
		return 0
	}

	// Remove extra whitespace
	text = strings.TrimSpace(text)

	// Count characters and estimate tokens
	// English text is roughly 4 characters per token
	// Code is roughly 3-3.5 characters per token
	// We'll use 3.5 as a conservative estimate
	charCount := len(text)
	estimatedTokens := int(float64(charCount) / 3.5)

	// Add some tokens for special tokens and formatting
	estimatedTokens += 10

	// Minimum of 1 token for non-empty text
	if estimatedTokens < 1 {
		estimatedTokens = 1
	}

	return estimatedTokens
}

// TruncateToTokenLimit truncates text to fit within token limits
func (c *OpenAIClient) TruncateToTokenLimit(text string, maxTokens int) (string, error) {
	currentTokens, err := c.CountTokens(text)
	if err != nil {
		return "", err
	}

	if currentTokens <= maxTokens {
		return text, nil
	}

	// Calculate approximate character limit
	// Use a conservative ratio to ensure we don't exceed token limit
	charRatio := 3.0 // Conservative estimate: 3 chars per token
	maxChars := int(float64(maxTokens) * charRatio)

	if maxChars >= len(text) {
		return text, nil
	}

	// Truncate at word boundary if possible
	truncated := text[:maxChars]

	// Find the last space to avoid cutting words
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxChars/2 { // Only use word boundary if it's not too far back
		truncated = truncated[:lastSpace]
	}

	// Add ellipsis to indicate truncation
	if len(truncated) < len(text) {
		truncated += "..."
	}

	return truncated, nil
}

// ValidateTokenLimits checks if the request would exceed model limits
func (c *OpenAIClient) ValidateTokenLimits(messages []Message) error {
	totalTokens := 0

	for _, message := range messages {
		tokens, err := c.CountTokens(message.Content)
		if err != nil {
			return fmt.Errorf("failed to count tokens for message: %w", err)
		}
		totalTokens += tokens
		totalTokens += 4 // Account for message formatting tokens
	}

	// Add tokens for the response
	totalTokens += c.config.MaxTokens

	// Check against model limits
	modelLimit := c.getModelTokenLimit()
	if totalTokens > modelLimit {
		return fmt.Errorf("request would exceed model token limit: %d > %d", totalTokens, modelLimit)
	}

	return nil
}

// getModelTokenLimit returns the context window size for the current model
func (c *OpenAIClient) getModelTokenLimit() int {
	model := strings.ToLower(c.config.Model)

	switch {
	case strings.Contains(model, "gpt-4o"):
		return 128000 // 128k context window
	case strings.Contains(model, "gpt-4-turbo") || strings.Contains(model, "gpt-4-1106") || strings.Contains(model, "gpt-4-0125"):
		return 128000 // 128k context window
	case strings.Contains(model, "gpt-4"):
		return 8192 // 8k context window for older GPT-4
	case strings.Contains(model, "gpt-3.5-turbo-16k"):
		return 16384 // 16k context window
	case strings.Contains(model, "gpt-3.5-turbo"):
		return 4096 // 4k context window
	default:
		return 128000 // Default to modern large context window
	}
}

// FormatUsageReport creates a human-readable usage report
func (c *OpenAIClient) FormatUsageReport() string {
	stats := c.GetUsageStats()

	return fmt.Sprintf(`Token Usage Report:
  Prompt Tokens: %d
  Completion Tokens: %d
  Total Tokens: %d
  Estimated Cost: $%.4f`,
		stats.PromptTokens,
		stats.CompletionTokens,
		stats.TotalTokens,
		stats.EstimatedCost,
	)
}
