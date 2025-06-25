package embeddings

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/processor"
)

func TestDefaultEmbeddingConfig(t *testing.T) {
	config := DefaultEmbeddingConfig()

	if config.Model != "text-embedding-3-small" {
		t.Errorf("Expected model text-embedding-3-small, got %s", config.Model)
	}
	if config.Dimensions != 1536 {
		t.Errorf("Expected dimensions 1536, got %d", config.Dimensions)
	}
	if config.BatchSize != 100 {
		t.Errorf("Expected batch size 100, got %d", config.BatchSize)
	}
	if config.StoragePath != "./embeddings.db" {
		t.Errorf("Expected storage path ./embeddings.db, got %s", config.StoragePath)
	}
}

func TestDefaultVectorSearchOptions(t *testing.T) {
	options := DefaultVectorSearchOptions()

	if options.TopK != 20 {
		t.Errorf("Expected TopK 20, got %d", options.TopK)
	}
	if options.MinScore != 0.1 {
		t.Errorf("Expected MinScore 0.1, got %f", options.MinScore)
	}
	if !options.IncludeContent {
		t.Error("Expected IncludeContent to be true")
	}
}

func TestSimilarityCalculator(t *testing.T) {
	calc := NewSimilarityCalculator(CosineSimilarity)

	// Test cosine similarity
	v1 := []float32{1.0, 0.0, 0.0}
	v2 := []float32{1.0, 0.0, 0.0}
	similarity := calc.Calculate(v1, v2)

	// For identical vectors, cosine similarity should be high
	if similarity <= 0 {
		t.Errorf("Expected positive similarity for identical vectors, got %f", similarity)
	}

	// Test with different vectors
	v3 := []float32{0.0, 1.0, 0.0}
	similarity2 := calc.Calculate(v1, v3)

	// For orthogonal vectors, similarity should be lower
	if similarity2 >= similarity {
		t.Errorf("Expected lower similarity for orthogonal vectors")
	}

	// Test with different length vectors (should return 0)
	v4 := []float32{1.0, 0.0}
	similarity3 := calc.Calculate(v1, v4)
	if similarity3 != 0.0 {
		t.Errorf("Expected 0 similarity for different length vectors, got %f", similarity3)
	}
}

func TestSimilarityMetrics(t *testing.T) {
	v1 := []float32{1.0, 2.0, 3.0}
	v2 := []float32{2.0, 3.0, 4.0}

	tests := []struct {
		metric   SimilarityMetric
		expected bool // Just check if it returns a reasonable value
	}{
		{CosineSimilarity, true},
		{EuclideanDistance, true},
		{DotProduct, true},
		{ManhattanDistance, true},
	}

	for _, test := range tests {
		calc := NewSimilarityCalculator(test.metric)
		result := calc.Calculate(v1, v2)

		// Just verify we get a numeric result (not NaN or infinite)
		if result != result { // NaN check
			t.Errorf("Metric %s returned NaN", test.metric)
		}
	}
}

func TestBoltVectorDB(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := &EmbeddingConfig{
		StoragePath: dbPath,
		Dimensions:  3,
		Timeout:     30,
	}

	db, err := NewBoltVectorDB(config)
	if err != nil {
		t.Fatalf("Failed to create vector database: %v", err)
	}
	defer db.Close()

	// Test storing document embedding
	embedding := &DocumentEmbedding{
		DocumentID:  "test-doc-1",
		FilePath:    "test.go",
		Language:    "Go",
		Category:    "code",
		ChunkCount:  2,
		ProcessedAt: time.Now(),
		Embeddings: []EmbeddingVector{
			{
				ID:        "chunk-1",
				Vector:    []float32{1.0, 0.0, 0.0},
				Content:   "test content chunk 1",
				Dimension: 3,
				Metadata: map[string]string{
					"chunkId": "chunk-1",
				},
				CreatedAt: time.Now(),
			},
			{
				ID:        "chunk-2",
				Vector:    []float32{0.0, 1.0, 0.0},
				Content:   "test content chunk 2",
				Dimension: 3,
				Metadata: map[string]string{
					"chunkId": "chunk-2",
				},
				CreatedAt: time.Now(),
			},
		},
	}

	// Test Store
	err = db.Store(embedding)
	if err != nil {
		t.Errorf("Failed to store embedding: %v", err)
	}

	// Test Get
	retrieved, err := db.Get("test-doc-1")
	if err != nil {
		t.Errorf("Failed to get embedding: %v", err)
	}
	if retrieved.DocumentID != "test-doc-1" {
		t.Errorf("Expected document ID test-doc-1, got %s", retrieved.DocumentID)
	}
	if len(retrieved.Embeddings) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(retrieved.Embeddings))
	}

	// Test List
	docIDs, err := db.List()
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}
	if len(docIDs) != 1 {
		t.Errorf("Expected 1 document, got %d", len(docIDs))
	}
	if docIDs[0] != "test-doc-1" {
		t.Errorf("Expected document ID test-doc-1, got %s", docIDs[0])
	}

	// Test Search
	queryVector := []float32{1.0, 0.0, 0.0}
	searchOptions := &VectorSearchOptions{
		TopK:           10,
		MinScore:       0.1,
		IncludeContent: true,
	}

	results, err := db.Search(queryVector, searchOptions)
	if err != nil {
		t.Errorf("Failed to search: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one search result")
	}

	// Test GetStats
	stats, err := db.GetStats()
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}
	if stats.TotalDocuments != 1 {
		t.Errorf("Expected 1 document in stats, got %d", stats.TotalDocuments)
	}
	if stats.TotalEmbeddings != 2 {
		t.Errorf("Expected 2 embeddings in stats, got %d", stats.TotalEmbeddings)
	}

	// Test Delete
	err = db.Delete("test-doc-1")
	if err != nil {
		t.Errorf("Failed to delete document: %v", err)
	}

	// Verify deletion
	_, err = db.Get("test-doc-1")
	if err == nil {
		t.Error("Expected error when getting deleted document")
	}
}

func TestBoltVectorDBBatch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_batch.db")

	config := &EmbeddingConfig{
		StoragePath: dbPath,
		Dimensions:  3,
		Timeout:     30,
	}

	db, err := NewBoltVectorDB(config)
	if err != nil {
		t.Fatalf("Failed to create vector database: %v", err)
	}
	defer db.Close()

	// Create multiple embeddings
	embeddings := []*DocumentEmbedding{
		{
			DocumentID:  "doc-1",
			FilePath:    "file1.go",
			Language:    "Go",
			Category:    "code",
			ChunkCount:  1,
			ProcessedAt: time.Now(),
			Embeddings: []EmbeddingVector{
				{
					ID:        "chunk-1-1",
					Vector:    []float32{1.0, 0.0, 0.0},
					Content:   "Go code content",
					Dimension: 3,
					Metadata: map[string]string{
						"language": "Go",
					},
					CreatedAt: time.Now(),
				},
			},
		},
		{
			DocumentID:  "doc-2",
			FilePath:    "file2.py",
			Language:    "Python",
			Category:    "code",
			ChunkCount:  1,
			ProcessedAt: time.Now(),
			Embeddings: []EmbeddingVector{
				{
					ID:        "chunk-2-1",
					Vector:    []float32{0.0, 1.0, 0.0},
					Content:   "Python code content",
					Dimension: 3,
					Metadata: map[string]string{
						"language": "Python",
					},
					CreatedAt: time.Now(),
				},
			},
		},
	}

	// Test StoreBatch
	err = db.StoreBatch(embeddings)
	if err != nil {
		t.Errorf("Failed to store batch: %v", err)
	}

	// Verify all documents were stored
	docIDs, err := db.List()
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}
	if len(docIDs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docIDs))
	}

	// Test search with filter
	queryVector := []float32{1.0, 0.0, 0.0}
	searchOptions := &VectorSearchOptions{
		TopK:     10,
		MinScore: 0.1,
		FilterBy: map[string]string{
			"language": "Go",
		},
		IncludeContent: true,
	}

	results, err := db.Search(queryVector, searchOptions)
	if err != nil {
		t.Errorf("Failed to search with filter: %v", err)
	}

	// Should find at least the Go document
	foundGo := false
	for _, result := range results {
		if result.FilePath == "file1.go" {
			foundGo = true
		}
	}
	if !foundGo {
		t.Error("Expected to find Go document in filtered search")
	}
}

func TestEmbeddingService(t *testing.T) {
	// Create mock embedding generator
	mockGen := &MockEmbeddingGenerator{
		model:      "test-model",
		dimensions: 3,
		maxTokens:  1000,
	}

	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_service.db")

	config := &EmbeddingConfig{
		StoragePath: dbPath,
		Dimensions:  3,
		Timeout:     30,
	}

	db, err := NewBoltVectorDB(config)
	if err != nil {
		t.Fatalf("Failed to create vector database: %v", err)
	}
	defer db.Close()

	service := NewEmbeddingService(mockGen, db, config)

	// Create test document
	doc := processor.Document{
		ID:       "test-doc",
		FilePath: "test.go",
		Language: "Go",
		Category: "code",
		Chunks: []processor.TextChunk{
			{
				ID:   "chunk-1",
				Text: "func main() { fmt.Println(\"Hello\") }",
			},
			{
				ID:   "chunk-2",
				Text: "func add(a, b int) int { return a + b }",
			},
		},
	}

	// Test ProcessDocument
	embedding, err := service.ProcessDocument(doc)
	if err != nil {
		t.Errorf("Failed to process document: %v", err)
	}

	if embedding.DocumentID != doc.ID {
		t.Errorf("Expected document ID %s, got %s", doc.ID, embedding.DocumentID)
	}
	if len(embedding.Embeddings) != len(doc.Chunks) {
		t.Errorf("Expected %d embeddings, got %d", len(doc.Chunks), len(embedding.Embeddings))
	}

	// Test ProcessDocuments
	docs := []processor.Document{doc}
	err = service.ProcessDocuments(docs)
	if err != nil {
		t.Errorf("Failed to process documents: %v", err)
	}
}

// Mock embedding generator for testing
type MockEmbeddingGenerator struct {
	model      string
	dimensions int
	maxTokens  int
}

func (m *MockEmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	// Return a simple mock embedding
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(len(text)) / float32(100+i) // Simple hash-like function
	}
	return embedding, nil
}

func (m *MockEmbeddingGenerator) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := m.GenerateEmbedding(text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (m *MockEmbeddingGenerator) GetModel() string {
	return m.model
}

func (m *MockEmbeddingGenerator) GetDimensions() int {
	return m.dimensions
}

func (m *MockEmbeddingGenerator) GetMaxTokens() int {
	return m.maxTokens
}

func (m *MockEmbeddingGenerator) EstimateTokens(text string) int {
	return len(text) / 4 // Simple estimation
}

func (m *MockEmbeddingGenerator) SplitTextForEmbedding(text string, maxTokens int) []string {
	if m.EstimateTokens(text) <= maxTokens {
		return []string{text}
	}
	// Simple splitting
	mid := len(text) / 2
	return []string{text[:mid], text[mid:]}
}

func TestVectorSearchFiltering(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_filtering.db")

	config := &EmbeddingConfig{
		StoragePath: dbPath,
		Dimensions:  3,
	}

	db, err := NewBoltVectorDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store test embeddings with different metadata
	embeddings := []*DocumentEmbedding{
		{
			DocumentID: "go-doc",
			FilePath:   "main.go",
			Language:   "Go",
			Category:   "code",
			ChunkCount: 1,
			Embeddings: []EmbeddingVector{
				{
					ID:      "go-chunk",
					Vector:  []float32{1.0, 0.0, 0.0},
					Content: "func main() { fmt.Println(\"Hello\") }",
					Metadata: map[string]string{
						"language": "Go",
						"type":     "function",
					},
				},
			},
		},
		{
			DocumentID: "py-doc",
			FilePath:   "main.py",
			Language:   "Python",
			Category:   "code",
			ChunkCount: 1,
			Embeddings: []EmbeddingVector{
				{
					ID:      "py-chunk",
					Vector:  []float32{0.0, 1.0, 0.0},
					Content: "class TestClass: pass",
					Metadata: map[string]string{
						"language": "Python",
						"type":     "class",
					},
				},
			},
		},
	}

	err = db.StoreBatch(embeddings)
	if err != nil {
		t.Fatalf("Failed to store embeddings: %v", err)
	}

	// Test filtering by language
	searchOptions := &VectorSearchOptions{
		TopK:     10,
		MinScore: 0.0,
		FilterBy: map[string]string{
			"language": "Go",
		},
	}

	results, err := db.Search([]float32{1.0, 0.0, 0.0}, searchOptions)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Should only return Go results
	if len(results) != 1 {
		t.Errorf("Expected 1 result with Go filter, got %d", len(results))
	}
	if len(results) > 0 && results[0].DocumentID != "go-doc" {
		t.Errorf("Expected go-doc, got %s", results[0].DocumentID)
	}
}

// Benchmark tests
func BenchmarkVectorSearch(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	config := &EmbeddingConfig{
		StoragePath: dbPath,
		Dimensions:  128,
	}

	db, err := NewBoltVectorDB(config)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create many test embeddings
	embeddings := make([]*DocumentEmbedding, 100)
	for i := 0; i < 100; i++ {
		vector := make([]float32, 128)
		for j := range vector {
			vector[j] = float32(i+j) / 100.0
		}

		embeddings[i] = &DocumentEmbedding{
			DocumentID: "doc-" + string(rune(i)),
			FilePath:   "file" + string(rune(i)) + ".go",
			Language:   "Go",
			Category:   "code",
			ChunkCount: 1,
			Embeddings: []EmbeddingVector{
				{
					ID:      "chunk-" + string(rune(i)),
					Vector:  vector,
					Content: "test content " + string(rune(i)),
				},
			},
		}
	}

	err = db.StoreBatch(embeddings)
	if err != nil {
		b.Fatalf("Failed to store embeddings: %v", err)
	}

	queryVector := make([]float32, 128)
	for i := range queryVector {
		queryVector[i] = 0.5
	}

	searchOptions := &VectorSearchOptions{
		TopK:     10,
		MinScore: 0.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Search(queryVector, searchOptions)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}
