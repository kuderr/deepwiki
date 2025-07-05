package embeddings

import (
	"errors"
	"time"

	"github.com/kuderr/deepwiki/pkg/processor"
)

// Common errors
var (
	ErrNotFound      = errors.New("document not found")
	ErrInvalidVector = errors.New("invalid vector dimensions")
	ErrInvalidConfig = errors.New("invalid configuration")
)

// EmbeddingVector represents an embedding vector with metadata
type EmbeddingVector struct {
	ID        string            `json:"id"`        // Unique identifier
	Vector    []float32         `json:"vector"`    // Embedding vector
	Content   string            `json:"content"`   // Original chunk content
	Dimension int               `json:"dimension"` // Vector dimension
	Metadata  map[string]string `json:"metadata"`  // Additional metadata
	CreatedAt time.Time         `json:"createdAt"` // When embedding was created
}

// DocumentEmbedding represents embeddings for a complete document
type DocumentEmbedding struct {
	DocumentID  string            `json:"documentId"`  // Reference to original document
	FilePath    string            `json:"filePath"`    // Original file path
	Language    string            `json:"language"`    // Programming language
	Category    string            `json:"category"`    // File category
	ChunkCount  int               `json:"chunkCount"`  // Number of chunks
	Embeddings  []EmbeddingVector `json:"embeddings"`  // Chunk embeddings
	Summary     *EmbeddingVector  `json:"summary"`     // Optional document-level embedding
	ProcessedAt time.Time         `json:"processedAt"` // When embeddings were created
}

// VectorSearchResult represents a search result with similarity score
type VectorSearchResult struct {
	DocumentID string            `json:"documentId"` // Document identifier
	ChunkID    string            `json:"chunkId"`    // Chunk identifier
	FilePath   string            `json:"filePath"`   // File path
	Content    string            `json:"content"`    // Chunk content
	Score      float32           `json:"score"`      // Similarity score
	Metadata   map[string]string `json:"metadata"`   // Additional metadata
}

// VectorSearchOptions represents options for vector search
type VectorSearchOptions struct {
	TopK           int                `json:"topK"`           // Number of results to return
	MinScore       float32            `json:"minScore"`       // Minimum similarity score
	MaxResults     int                `json:"maxResults"`     // Maximum results (different from TopK for reranking)
	FilterBy       map[string]string  `json:"filterBy"`       // Metadata filters
	BoostFactors   map[string]float32 `json:"boostFactors"`   // Boost factors for different attributes
	IncludeContent bool               `json:"includeContent"` // Whether to include full content
}

// DefaultVectorSearchOptions returns default search options
func DefaultVectorSearchOptions() *VectorSearchOptions {
	return &VectorSearchOptions{
		TopK:           20,
		MinScore:       0.1,
		MaxResults:     50,
		FilterBy:       make(map[string]string),
		BoostFactors:   make(map[string]float32),
		IncludeContent: true,
	}
}

// EmbeddingConfig represents configuration for embeddings
type EmbeddingConfig struct {
	// OpenAI settings
	Model      string `json:"model"`      // Embedding model (e.g., "text-embedding-3-small")
	Dimensions int    `json:"dimensions"` // Vector dimensions
	BatchSize  int    `json:"batchSize"`  // Batch size for API calls

	// Processing settings
	ChunkSize  int `json:"chunkSize"`  // Max tokens per chunk for embedding
	MaxRetries int `json:"maxRetries"` // Max retries for API calls
	Timeout    int `json:"timeout"`    // Timeout in seconds

	// Storage settings
	StoragePath string `json:"storagePath"` // Path to vector database file
	Compress    bool   `json:"compress"`    // Whether to compress vectors
}

// DefaultEmbeddingConfig returns default embedding configuration
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		Model:       "text-embedding-3-small",
		Dimensions:  1536, // Default for text-embedding-3-small
		BatchSize:   100,
		ChunkSize:   8000, // Max tokens for embedding
		MaxRetries:  3,
		Timeout:     30,
		StoragePath: "./embeddings.db",
		Compress:    false,
	}
}

// VectorDatabase interface defines operations for vector storage
type VectorDatabase interface {
	// Storage operations
	Store(embedding *DocumentEmbedding) error
	StoreBatch(embeddings []*DocumentEmbedding) error
	Get(documentID string) (*DocumentEmbedding, error)
	Delete(documentID string) error
	List() ([]string, error) // Returns list of document IDs

	// Search operations
	Search(vector []float32, options *VectorSearchOptions) ([]VectorSearchResult, error)

	// Maintenance operations
	Optimize() error
	GetStats() (*DatabaseStats, error)
	Close() error
}

// DatabaseStats represents statistics about the vector database
type DatabaseStats struct {
	TotalDocuments   int       `json:"totalDocuments"`   // Total number of documents
	TotalEmbeddings  int       `json:"totalEmbeddings"`  // Total number of embeddings
	TotalSize        int64     `json:"totalSize"`        // Database size in bytes
	IndexSize        int64     `json:"indexSize"`        // Index size in bytes
	LastOptimized    time.Time `json:"lastOptimized"`    // Last optimization time
	AverageDimension int       `json:"averageDimension"` // Average vector dimension
	CreatedAt        time.Time `json:"createdAt"`        // Database creation time
	UpdatedAt        time.Time `json:"updatedAt"`        // Last update time
}

// EmbeddingGenerator interface for generating embeddings
type EmbeddingGenerator interface {
	// Generate embeddings for text chunks
	GenerateEmbedding(text string) ([]float32, error)
	GenerateBatchEmbeddings(texts []string) ([][]float32, error)

	// Get model information
	GetModel() string
	GetDimensions() int
	GetMaxTokens() int

	// Utilities
	EstimateTokens(text string) int
	SplitTextForEmbedding(text string, maxTokens int) []string
}

// SimilarityMetric represents different similarity calculation methods
type SimilarityMetric string

const (
	CosineSimilarity  SimilarityMetric = "cosine"    // Cosine similarity
	EuclideanDistance SimilarityMetric = "euclidean" // Euclidean distance
	DotProduct        SimilarityMetric = "dot"       // Dot product
	ManhattanDistance SimilarityMetric = "manhattan" // Manhattan distance
)

// SimilarityCalculator provides methods for calculating vector similarities
type SimilarityCalculator struct {
	metric SimilarityMetric
}

// NewSimilarityCalculator creates a new similarity calculator
func NewSimilarityCalculator(metric SimilarityMetric) *SimilarityCalculator {
	return &SimilarityCalculator{metric: metric}
}

// Calculate computes similarity between two vectors
func (sc *SimilarityCalculator) Calculate(v1, v2 []float32) float32 {
	if len(v1) != len(v2) {
		return 0.0
	}

	switch sc.metric {
	case CosineSimilarity:
		return sc.cosineSimilarity(v1, v2)
	case EuclideanDistance:
		return 1.0 / (1.0 + sc.euclideanDistance(v1, v2)) // Convert distance to similarity
	case DotProduct:
		return sc.dotProduct(v1, v2)
	case ManhattanDistance:
		return 1.0 / (1.0 + sc.manhattanDistance(v1, v2)) // Convert distance to similarity
	default:
		return sc.cosineSimilarity(v1, v2)
	}
}

// cosineSimilarity calculates cosine similarity between vectors
func (sc *SimilarityCalculator) cosineSimilarity(v1, v2 []float32) float32 {
	var dotProduct, normA, normB float32

	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		normA += v1[i] * v1[i]
		normB += v2[i] * v2[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (float32(len(v1)) * float32(len(v2)))
}

// euclideanDistance calculates Euclidean distance between vectors
func (sc *SimilarityCalculator) euclideanDistance(v1, v2 []float32) float32 {
	var sum float32
	for i := 0; i < len(v1); i++ {
		diff := v1[i] - v2[i]
		sum += diff * diff
	}
	return float32(len(v1)) // Simplified for now
}

// dotProduct calculates dot product of two vectors
func (sc *SimilarityCalculator) dotProduct(v1, v2 []float32) float32 {
	var product float32
	for i := 0; i < len(v1); i++ {
		product += v1[i] * v2[i]
	}
	return product
}

// manhattanDistance calculates Manhattan distance between vectors
func (sc *SimilarityCalculator) manhattanDistance(v1, v2 []float32) float32 {
	var sum float32
	for i := 0; i < len(v1); i++ {
		if v1[i] > v2[i] {
			sum += v1[i] - v2[i]
		} else {
			sum += v2[i] - v1[i]
		}
	}
	return sum
}

// EmbeddingService coordinates embedding generation and storage
type EmbeddingService struct {
	generator  EmbeddingGenerator
	database   VectorDatabase
	calculator *SimilarityCalculator
	config     *EmbeddingConfig
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(
	generator EmbeddingGenerator,
	database VectorDatabase,
	config *EmbeddingConfig,
) *EmbeddingService {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	return &EmbeddingService{
		generator:  generator,
		database:   database,
		calculator: NewSimilarityCalculator(CosineSimilarity),
		config:     config,
	}
}

// ProcessDocuments generates and stores embeddings for documents
func (es *EmbeddingService) ProcessDocuments(documents []processor.Document) error {
	embeddings := make([]*DocumentEmbedding, 0, len(documents))

	for _, doc := range documents {
		embedding, err := es.ProcessDocument(doc)
		if err != nil {
			return err
		}
		embeddings = append(embeddings, embedding)
	}

	return es.database.StoreBatch(embeddings)
}

// ProcessDocument generates embeddings for a single document
func (es *EmbeddingService) ProcessDocument(doc processor.Document) (*DocumentEmbedding, error) {
	// Prepare texts for embedding
	texts := make([]string, len(doc.Chunks))
	for i, chunk := range doc.Chunks {
		texts[i] = chunk.Text
	}

	// Generate embeddings
	vectors, err := es.generator.GenerateBatchEmbeddings(texts)
	if err != nil {
		return nil, err
	}

	// Create embedding vectors
	embeddingVectors := make([]EmbeddingVector, len(vectors))
	for i, vector := range vectors {
		embeddingVectors[i] = EmbeddingVector{
			ID:        doc.Chunks[i].ID,
			Vector:    vector,
			Content:   doc.Chunks[i].Text, // Store the original chunk content
			Dimension: len(vector),
			Metadata: map[string]string{
				"chunkId":   doc.Chunks[i].ID,
				"wordCount": string(rune(doc.Chunks[i].WordCount)),
				"startPos":  string(rune(doc.Chunks[i].StartPos)),
				"endPos":    string(rune(doc.Chunks[i].EndPos)),
			},
			CreatedAt: time.Now(),
		}
	}

	return &DocumentEmbedding{
		DocumentID:  doc.ID,
		FilePath:    doc.FilePath,
		Language:    doc.Language,
		Category:    doc.Category,
		ChunkCount:  len(doc.Chunks),
		Embeddings:  embeddingVectors,
		ProcessedAt: time.Now(),
	}, nil
}
