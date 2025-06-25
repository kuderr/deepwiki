package embeddings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/openai"
)

// OpenAIEmbeddingGenerator implements EmbeddingGenerator using OpenAI API
type OpenAIEmbeddingGenerator struct {
	client openai.Client
	config *EmbeddingConfig
}

// NewOpenAIEmbeddingGenerator creates a new OpenAI embedding generator
func NewOpenAIEmbeddingGenerator(client openai.Client, config *EmbeddingConfig) *OpenAIEmbeddingGenerator {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	return &OpenAIEmbeddingGenerator{
		client: client,
		config: config,
	}
}

// GenerateEmbedding generates an embedding for a single text
func (g *OpenAIEmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return nil, fmt.Errorf("empty text provided")
	}

	// Check token count
	tokenCount := g.EstimateTokens(text)
	if tokenCount > g.config.ChunkSize {
		// Split text if too large
		chunks := g.SplitTextForEmbedding(text, g.config.ChunkSize)
		if len(chunks) == 0 {
			return nil, fmt.Errorf("text too large and cannot be split")
		}
		// Use first chunk for now - in practice, you might want to average embeddings
		text = chunks[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.config.Timeout)*time.Second)
	defer cancel()

	// Create embedding request - note: single text needs to be in a slice
	texts := []string{text}

	response, err := g.client.CreateEmbeddings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	// Convert from []float64 to []float32
	embedding64 := response.Data[0].Embedding
	embedding32 := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding32[i] = float32(v)
	}

	return embedding32, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (g *OpenAIEmbeddingGenerator) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Filter out empty texts
	validTexts := make([]string, 0, len(texts))
	textIndexMap := make(map[int]int) // Maps result index to original index

	for i, text := range texts {
		if len(strings.TrimSpace(text)) > 0 {
			// Check and split if necessary
			if g.EstimateTokens(text) > g.config.ChunkSize {
				chunks := g.SplitTextForEmbedding(text, g.config.ChunkSize)
				if len(chunks) > 0 {
					validTexts = append(validTexts, chunks[0])
					textIndexMap[len(validTexts)-1] = i
				}
			} else {
				validTexts = append(validTexts, text)
				textIndexMap[len(validTexts)-1] = i
			}
		}
	}

	if len(validTexts) == 0 {
		return nil, fmt.Errorf("no valid texts to process")
	}

	// Process in batches
	allEmbeddings := make([][]float32, len(texts))
	batchSize := g.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(validTexts); i += batchSize {
		end := i + batchSize
		if end > len(validTexts) {
			end = len(validTexts)
		}

		batch := validTexts[i:end]
		embeddings, err := g.processBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to process batch %d-%d: %v", i, end, err)
		}

		// Map embeddings back to original indices
		for j, embedding := range embeddings {
			originalIndex := textIndexMap[i+j]
			allEmbeddings[originalIndex] = embedding
		}
	}

	return allEmbeddings, nil
}

// processBatch processes a single batch of texts
func (g *OpenAIEmbeddingGenerator) processBatch(texts []string) ([][]float32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.config.Timeout)*time.Second)
	defer cancel()

	var response *openai.EmbeddingResponse
	var err error

	// Retry logic
	for attempt := 0; attempt <= g.config.MaxRetries; attempt++ {
		response, err = g.client.CreateEmbeddings(ctx, texts)
		if err == nil {
			break
		}

		if attempt < g.config.MaxRetries {
			// Exponential backoff
			backoff := time.Duration(1<<attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings after %d attempts: %v", g.config.MaxRetries+1, err)
	}

	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(response.Data))
	}

	// Extract embeddings and convert from []float64 to []float32
	embeddings := make([][]float32, len(response.Data))
	for i, data := range response.Data {
		embedding64 := data.Embedding
		embedding32 := make([]float32, len(embedding64))
		for j, v := range embedding64 {
			embedding32[j] = float32(v)
		}
		embeddings[i] = embedding32
	}

	return embeddings, nil
}

// GetModel returns the embedding model name
func (g *OpenAIEmbeddingGenerator) GetModel() string {
	return g.config.Model
}

// GetDimensions returns the embedding dimensions
func (g *OpenAIEmbeddingGenerator) GetDimensions() int {
	return g.config.Dimensions
}

// GetMaxTokens returns the maximum tokens for embedding
func (g *OpenAIEmbeddingGenerator) GetMaxTokens() int {
	return g.config.ChunkSize
}

// EstimateTokens estimates the number of tokens in text
func (g *OpenAIEmbeddingGenerator) EstimateTokens(text string) int {
	// Use the OpenAI client's token counting if available
	if g.client != nil {
		if count, err := g.client.CountTokens(text); err == nil {
			return count
		}
	}

	// TODO: Implement fallback token counting using tiktoken-go when client is unavailable
	return len(text) / 4
}

// SplitTextForEmbedding splits text into chunks that fit within token limits
func (g *OpenAIEmbeddingGenerator) SplitTextForEmbedding(text string, maxTokens int) []string {
	if g.EstimateTokens(text) <= maxTokens {
		return []string{text}
	}

	chunks := make([]string, 0)

	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")
	currentChunk := ""

	for _, paragraph := range paragraphs {
		testChunk := currentChunk
		if len(testChunk) > 0 {
			testChunk += "\n\n"
		}
		testChunk += paragraph

		if g.EstimateTokens(testChunk) <= maxTokens {
			currentChunk = testChunk
		} else {
			// Save current chunk if it has content
			if len(currentChunk) > 0 {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			// If single paragraph is too large, split by sentences
			if g.EstimateTokens(paragraph) > maxTokens {
				sentenceChunks := g.splitBySentences(paragraph, maxTokens)
				chunks = append(chunks, sentenceChunks...)
			} else {
				currentChunk = paragraph
			}
		}
	}

	// Add remaining content
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

// splitBySentences splits text by sentences to fit within token limits
func (g *OpenAIEmbeddingGenerator) splitBySentences(text string, maxTokens int) []string {
	// Simple sentence splitting by periods, exclamation marks, and question marks
	sentences := strings.FieldsFunc(text, func(c rune) bool {
		return c == '.' || c == '!' || c == '?'
	})

	chunks := make([]string, 0)
	currentChunk := ""

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) == 0 {
			continue
		}

		testChunk := currentChunk
		if len(testChunk) > 0 {
			testChunk += ". "
		}
		testChunk += sentence

		if g.EstimateTokens(testChunk) <= maxTokens {
			currentChunk = testChunk
		} else {
			if len(currentChunk) > 0 {
				chunks = append(chunks, currentChunk+".")
				currentChunk = sentence
			} else {
				// Single sentence is too large, split by words
				wordChunks := g.splitByWords(sentence, maxTokens)
				chunks = append(chunks, wordChunks...)
			}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk+".")
	}

	return chunks
}

// splitByWords splits text by words to fit within token limits
func (g *OpenAIEmbeddingGenerator) splitByWords(text string, maxTokens int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	chunks := make([]string, 0)
	currentChunk := make([]string, 0)

	for _, word := range words {
		testChunk := append(currentChunk, word)
		testText := strings.Join(testChunk, " ")

		if g.EstimateTokens(testText) <= maxTokens {
			currentChunk = testChunk
		} else {
			if len(currentChunk) > 0 {
				chunks = append(chunks, strings.Join(currentChunk, " "))
				currentChunk = []string{word}
			} else {
				// TODO: Implement proper word truncation with character boundary respect
				chunks = append(chunks, word[:maxTokens*4]) // Rough truncation
			}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

// ValidateConfig validates the embedding configuration
func (g *OpenAIEmbeddingGenerator) ValidateConfig() error {
	if g.config.Model == "" {
		return fmt.Errorf("embedding model not specified")
	}

	if g.config.Dimensions <= 0 {
		return fmt.Errorf("invalid dimensions: %d", g.config.Dimensions)
	}

	if g.config.ChunkSize <= 0 {
		return fmt.Errorf("invalid chunk size: %d", g.config.ChunkSize)
	}

	if g.config.BatchSize <= 0 {
		return fmt.Errorf("invalid batch size: %d", g.config.BatchSize)
	}

	// Validate model-specific constraints
	switch g.config.Model {
	case "text-embedding-3-small":
		if g.config.Dimensions > 1536 {
			return fmt.Errorf("text-embedding-3-small supports max 1536 dimensions, got %d", g.config.Dimensions)
		}
	case "text-embedding-3-large":
		if g.config.Dimensions > 3072 {
			return fmt.Errorf("text-embedding-3-large supports max 3072 dimensions, got %d", g.config.Dimensions)
		}
	case "text-embedding-ada-002":
		if g.config.Dimensions != 1536 {
			return fmt.Errorf("text-embedding-ada-002 only supports 1536 dimensions, got %d", g.config.Dimensions)
		}
	default:
		// Unknown model, use defaults
	}

	return nil
}

// GetModelInfo returns information about the current model
func (g *OpenAIEmbeddingGenerator) GetModelInfo() map[string]interface{} {
	info := map[string]interface{}{
		"model":      g.config.Model,
		"dimensions": g.config.Dimensions,
		"maxTokens":  g.config.ChunkSize,
		"batchSize":  g.config.BatchSize,
		"maxRetries": g.config.MaxRetries,
		"timeout":    g.config.Timeout,
	}

	// Add model-specific information
	switch g.config.Model {
	case "text-embedding-3-small":
		info["cost_per_1k_tokens"] = 0.00002
		info["max_dimensions"] = 1536
		info["performance"] = "high"
	case "text-embedding-3-large":
		info["cost_per_1k_tokens"] = 0.00013
		info["max_dimensions"] = 3072
		info["performance"] = "highest"
	case "text-embedding-ada-002":
		info["cost_per_1k_tokens"] = 0.0001
		info["max_dimensions"] = 1536
		info["performance"] = "standard"
	}

	return info
}
