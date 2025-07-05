package rag

import (
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/embeddings"
	"github.com/kuderr/deepwiki/pkg/processor"
)

func TestDefaultRAGConfig(t *testing.T) {
	config := DefaultRAGConfig()

	if config.DefaultMaxResults != 20 {
		t.Errorf("Expected DefaultMaxResults 20, got %d", config.DefaultMaxResults)
	}
	if config.RetrievalStrategy != QueryTypeHybrid {
		t.Errorf("Expected QueryTypeHybrid, got %s", config.RetrievalStrategy)
	}
	if !config.RerankResults {
		t.Error("Expected RerankResults to be true")
	}
	if config.SemanticWeight != 0.6 {
		t.Errorf("Expected SemanticWeight 0.6, got %f", config.SemanticWeight)
	}
}

func TestDocumentRetriever(t *testing.T) {
	// Create test documents
	docs := []processor.Document{
		{
			ID:       "doc1",
			FilePath: "main.go",
			Language: "Go",
			Category: "code",
			Content:  "package main\nfunc main() { fmt.Println(\"Hello\") }",
			Chunks: []processor.TextChunk{
				{
					ID:   "chunk1",
					Text: "package main",
				},
				{
					ID:   "chunk2",
					Text: "func main() { fmt.Println(\"Hello\") }",
				},
			},
		},
		{
			ID:       "doc2",
			FilePath: "utils.go",
			Language: "Go",
			Category: "code",
			Content:  "package utils\nfunc Add(a, b int) int { return a + b }",
			Chunks: []processor.TextChunk{
				{
					ID:   "chunk3",
					Text: "package utils",
				},
				{
					ID:   "chunk4",
					Text: "func Add(a, b int) int { return a + b }",
				},
			},
		},
		{
			ID:       "doc3",
			FilePath: "README.md",
			Language: "Markdown",
			Category: "docs",
			Content:  "# Test Project\nThis is a test project for Go.",
			Chunks: []processor.TextChunk{
				{
					ID:   "chunk5",
					Text: "# Test Project",
				},
				{
					ID:   "chunk6",
					Text: "This is a test project for Go.",
				},
			},
		},
	}

	// Create mock dependencies
	mockEmbGen := &MockEmbeddingGenerator{}
	mockVectorDB := &MockVectorDB{}

	config := DefaultRAGConfig()

	retriever := NewDocumentRetriever(nil, mockVectorDB, mockEmbGen, docs, config)

	// Test RetrieveByQuery
	results, err := retriever.RetrieveByQuery("main function", 5)
	if err != nil {
		t.Errorf("Failed to retrieve by query: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one result")
	}

	// Test RetrieveByTags
	results, err = retriever.RetrieveByTags([]string{"main", "function"}, 5)
	if err != nil {
		t.Errorf("Failed to retrieve by tags: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one result for tags")
	}

	// Test RetrieveCodeExamples
	results, err = retriever.RetrieveCodeExamples("Go", "function", 3)
	if err != nil {
		t.Errorf("Failed to retrieve code examples: %v", err)
	}

	// Should only return Go code results
	for _, result := range results {
		if result.Language != "Go" {
			t.Errorf("Expected Go language, got %s", result.Language)
		}
		if result.Category != "code" {
			t.Errorf("Expected code category, got %s", result.Category)
		}
	}

	// Test RetrieveDocumentation
	results, err = retriever.RetrieveDocumentation("test project", 3)
	if err != nil {
		t.Errorf("Failed to retrieve documentation: %v", err)
	}

	// Test FilterResults
	allResults := []RetrievalResult{
		{DocumentID: "doc1", Language: "Go", Category: "code"},
		{DocumentID: "doc2", Language: "Python", Category: "code"},
		{DocumentID: "doc3", Language: "Go", Category: "docs"},
	}

	filtered := retriever.FilterResults(allResults, map[string]string{
		"language": "Go",
	})

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(filtered))
	}

	for _, result := range filtered {
		if result.Language != "Go" {
			t.Errorf("Expected Go language in filtered results, got %s", result.Language)
		}
	}

	// Test GetRetrievalStats
	stats := retriever.GetRetrievalStats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}

func TestRetrievalContext(t *testing.T) {
	ctx := &RetrievalContext{
		Query:      "test query",
		QueryType:  QueryTypeSemantic,
		MaxResults: 10,
		MinScore:   0.5,
		Filters: map[string]string{
			"language": "Go",
		},
		TimeWindow: &TimeWindow{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
	}

	if ctx.Query != "test query" {
		t.Errorf("Expected query 'test query', got %s", ctx.Query)
	}
	if ctx.QueryType != QueryTypeSemantic {
		t.Errorf("Expected QueryTypeSemantic, got %s", ctx.QueryType)
	}
	if ctx.TimeWindow == nil {
		t.Error("Expected non-nil TimeWindow")
	}
}

func TestRetrievalStrategies(t *testing.T) {
	docs := []processor.Document{
		{
			ID:       "doc1",
			FilePath: "test.go",
			Language: "Go",
			Category: "code",
			Chunks: []processor.TextChunk{
				{
					ID:   "chunk1",
					Text: "func testFunction() { return true }",
				},
			},
		},
	}

	mockEmbGen := &MockEmbeddingGenerator{}
	mockVectorDB := &MockVectorDB{}
	config := DefaultRAGConfig()

	retriever := NewDocumentRetriever(nil, mockVectorDB, mockEmbGen, docs, config)

	testCases := []struct {
		queryType QueryType
		query     string
	}{
		{QueryTypeSemantic, "test function"},
		{QueryTypeKeyword, "testFunction"},
		{QueryTypeHybrid, "test function"},
		{QueryTypeStructural, "func"},
	}

	for _, tc := range testCases {
		ctx := &RetrievalContext{
			Query:      tc.query,
			QueryType:  tc.queryType,
			MaxResults: 5,
			MinScore:   0.1,
		}

		results, err := retriever.RetrieveRelevantDocuments(ctx)
		if err != nil {
			t.Errorf("Failed to retrieve with %s strategy: %v", tc.queryType, err)
		}

		// Should get some results for test data
		if len(results) == 0 && tc.queryType != QueryTypeSemantic {
			// Semantic might return 0 results due to mock vector DB
			t.Errorf("Expected results for %s strategy", tc.queryType)
		}
	}
}

func TestReranking(t *testing.T) {
	config := DefaultRAGConfig()
	config.RerankResults = true

	retriever := NewDocumentRetriever(nil, &MockVectorDB{}, &MockEmbeddingGenerator{}, nil, config)

	results := []RetrievalResult{
		{
			DocumentID: "doc1",
			ChunkID:    "chunk1",
			Content:    "func main() { fmt.Println(\"hello world\") }",
			Score:      0.5,
			Language:   "Go",
		},
		{
			DocumentID: "doc2",
			ChunkID:    "chunk2",
			Content:    "def hello(): print('hello world')",
			Score:      0.7,
			Language:   "Python",
		},
	}

	reranked, err := retriever.RerankResults(results, "hello world")
	if err != nil {
		t.Errorf("Failed to rerank results: %v", err)
	}

	if len(reranked) != len(results) {
		t.Errorf("Expected %d reranked results, got %d", len(results), len(reranked))
	}

	// Check that relevance info was added
	for i, result := range reranked {
		if result.Relevance.RelevanceScore == 0 {
			t.Errorf("Result %d should have relevance score", i)
		}
		if len(result.Relevance.MatchedTerms) == 0 {
			t.Errorf("Result %d should have matched terms", i)
		}
	}
}

func TestRelatedChunks(t *testing.T) {
	docs := []processor.Document{
		{
			ID:       "doc1",
			FilePath: "test.go",
			Language: "Go",
			Category: "code",
			Chunks: []processor.TextChunk{
				{
					ID:   "target-chunk",
					Text: "func calculateSum(a, b int) int { return a + b }",
				},
				{
					ID:   "related-chunk",
					Text: "func calculateProduct(a, b int) int { return a * b }",
				},
			},
		},
	}

	mockEmbGen := &MockEmbeddingGenerator{}
	mockVectorDB := &MockVectorDB{}
	config := DefaultRAGConfig()

	retriever := NewDocumentRetriever(nil, mockVectorDB, mockEmbGen, docs, config)

	// Test RetrieveRelatedChunks
	results, err := retriever.RetrieveRelatedChunks("target-chunk", 5)
	if err != nil {
		t.Errorf("Failed to retrieve related chunks: %v", err)
	}

	// Should not include the original chunk
	for _, result := range results {
		if result.ChunkID == "target-chunk" {
			t.Error("Related chunks should not include the original chunk")
		}
	}
}

func TestContextualRetrieval(t *testing.T) {
	docs := []processor.Document{
		{
			ID:       "doc1",
			FilePath: "math.go",
			Language: "Go",
			Category: "code",
			Chunks: []processor.TextChunk{
				{
					ID:   "chunk1",
					Text: "package math\nfunc addition(a, b int) int { return a + b }", // Changed Add to addition
				},
			},
		},
	}

	mockEmbGen := &MockEmbeddingGenerator{}
	mockVectorDB := &MockVectorDB{}
	config := DefaultRAGConfig()

	retriever := NewDocumentRetriever(nil, mockVectorDB, mockEmbGen, docs, config)

	// Provide context about math operations
	context := []RetrievalResult{
		{
			DocumentID: "doc1",
			Content:    "mathematical operations in Go",
		},
	}

	// Test RetrieveWithContext
	results, err := retriever.RetrieveWithContext("addition function", context, 5)
	if err != nil {
		t.Errorf("Failed to retrieve with context: %v", err)
	}

	// Should boost results related to the context
	if len(results) == 0 {
		t.Error("Expected results with contextual retrieval")
	}
}

// Mock implementations for testing

type MockEmbeddingGenerator struct{}

func (m *MockEmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	// Generate a simple mock embedding based on text length
	return []float32{float32(len(text)), 0.5, 0.3}, nil
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
	return "mock-model"
}

func (m *MockEmbeddingGenerator) GetDimensions() int {
	return 3
}

func (m *MockEmbeddingGenerator) GetMaxTokens() int {
	return 1000
}

func (m *MockEmbeddingGenerator) EstimateTokens(text string) int {
	return len(text) / 4
}

func (m *MockEmbeddingGenerator) SplitTextForEmbedding(text string, maxTokens int) []string {
	if m.EstimateTokens(text) <= maxTokens {
		return []string{text}
	}
	mid := len(text) / 2
	return []string{text[:mid], text[mid:]}
}

type MockVectorDB struct {
	embeddings map[string]*embeddings.DocumentEmbedding
}

func (m *MockVectorDB) Store(embedding *embeddings.DocumentEmbedding) error {
	if m.embeddings == nil {
		m.embeddings = make(map[string]*embeddings.DocumentEmbedding)
	}
	m.embeddings[embedding.DocumentID] = embedding
	return nil
}

func (m *MockVectorDB) StoreBatch(embeddings []*embeddings.DocumentEmbedding) error {
	for _, emb := range embeddings {
		if err := m.Store(emb); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockVectorDB) Get(documentID string) (*embeddings.DocumentEmbedding, error) {
	if m.embeddings == nil {
		return nil, embeddings.ErrNotFound
	}
	emb, exists := m.embeddings[documentID]
	if !exists {
		return nil, embeddings.ErrNotFound
	}
	return emb, nil
}

func (m *MockVectorDB) Delete(documentID string) error {
	if m.embeddings != nil {
		delete(m.embeddings, documentID)
	}
	return nil
}

func (m *MockVectorDB) List() ([]string, error) {
	if m.embeddings == nil {
		return nil, nil
	}
	ids := make([]string, 0, len(m.embeddings))
	for id := range m.embeddings {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MockVectorDB) Search(
	vector []float32,
	options *embeddings.VectorSearchOptions,
) ([]embeddings.VectorSearchResult, error) {
	// Return empty results for mock
	return []embeddings.VectorSearchResult{}, nil
}

func (m *MockVectorDB) SearchByText(
	text string,
	options *embeddings.VectorSearchOptions,
) ([]embeddings.VectorSearchResult, error) {
	return []embeddings.VectorSearchResult{}, nil
}

func (m *MockVectorDB) Optimize() error {
	return nil
}

func (m *MockVectorDB) GetStats() (*embeddings.DatabaseStats, error) {
	return &embeddings.DatabaseStats{
		TotalDocuments:  len(m.embeddings),
		TotalEmbeddings: len(m.embeddings),
	}, nil
}

func (m *MockVectorDB) Close() error {
	return nil
}

// Define missing error type
var ErrNotFound = embeddings.ErrNotFound

func TestCompleteWorkflow(t *testing.T) {
	// Test the complete workflow from documents to retrieval
	docs := []processor.Document{
		{
			ID:       "doc1",
			FilePath: "main.go",
			Language: "Go",
			Category: "code",
			Content:  "package main\nfunc main() { fmt.Println(\"Hello, World!\") }",
			Chunks: []processor.TextChunk{
				{
					ID:        "chunk1",
					Text:      "package main",
					WordCount: 2,
				},
				{
					ID:        "chunk2",
					Text:      "func main() { fmt.Println(\"Hello, World!\") }",
					WordCount: 7,
				},
			},
		},
	}

	// Create components
	config := DefaultRAGConfig()

	mockEmbGen := &MockEmbeddingGenerator{}
	mockVectorDB := &MockVectorDB{}

	retriever := NewDocumentRetriever(nil, mockVectorDB, mockEmbGen, docs, config)

	// Test different query types
	queryTypes := []QueryType{
		QueryTypeKeyword,
		QueryTypeHybrid,
		QueryTypeStructural,
	}

	for _, queryType := range queryTypes {
		ctx := &RetrievalContext{
			Query:      "main function",
			QueryType:  queryType,
			MaxResults: 5,
			MinScore:   0.1,
		}

		results, err := retriever.RetrieveRelevantDocuments(ctx)
		if err != nil {
			t.Errorf("Failed retrieval with %s: %v", queryType, err)
		}

		// Verify results structure
		for _, result := range results {
			if result.DocumentID == "" {
				t.Error("Result should have document ID")
			}
			if result.ChunkID == "" {
				t.Error("Result should have chunk ID")
			}
			if result.Content == "" {
				t.Error("Result should have content")
			}
			if result.Score < 0 {
				t.Error("Result should have non-negative score")
			}
		}
	}
}
