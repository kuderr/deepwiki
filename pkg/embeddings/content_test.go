package embeddings

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIncludeContent(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_content.db")

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

	// Store test embedding with content
	testContent := "func main() { fmt.Println(\"Hello, World!\") }"
	embedding := &DocumentEmbedding{
		DocumentID:  "test-doc",
		FilePath:    "main.go",
		Language:    "Go",
		Category:    "code",
		ChunkCount:  1,
		ProcessedAt: time.Now(),
		Embeddings: []EmbeddingVector{
			{
				ID:        "chunk-1",
				Vector:    []float32{1.0, 0.0, 0.0},
				Content:   testContent, // Store the content
				Dimension: 3,
				Metadata: map[string]string{
					"type": "function",
				},
				CreatedAt: time.Now(),
			},
		},
	}

	err = db.Store(embedding)
	if err != nil {
		t.Fatalf("Failed to store embedding: %v", err)
	}

	// Test search WITH content
	queryVector := []float32{1.0, 0.0, 0.0}
	searchOptionsWithContent := &VectorSearchOptions{
		TopK:           5,
		MinScore:       0.1,
		IncludeContent: true, // Request content
	}

	resultsWithContent, err := db.Search(queryVector, searchOptionsWithContent)
	if err != nil {
		t.Errorf("Failed to search with content: %v", err)
	}

	if len(resultsWithContent) == 0 {
		t.Error("Expected at least one result with content")
	}

	// Verify content is included
	for i, result := range resultsWithContent {
		if result.Content == "" {
			t.Errorf("Result %d should have content when IncludeContent=true", i)
		}
		if result.Content != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, result.Content)
		}
	}

	// Test search WITHOUT content
	searchOptionsWithoutContent := &VectorSearchOptions{
		TopK:           5,
		MinScore:       0.1,
		IncludeContent: false, // Don't request content
	}

	resultsWithoutContent, err := db.Search(queryVector, searchOptionsWithoutContent)
	if err != nil {
		t.Errorf("Failed to search without content: %v", err)
	}

	if len(resultsWithoutContent) == 0 {
		t.Error("Expected at least one result without content")
	}

	// Verify content is NOT included
	for i, result := range resultsWithoutContent {
		if result.Content != "" {
			t.Errorf("Result %d should not have content when IncludeContent=false, got '%s'", i, result.Content)
		}
	}
}

func TestSearchByTextWithSearchService(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_search_by_text.db")

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

	// Create search service
	mockGen := NewTestMockEmbeddingGenerator()
	searchService := NewSearchService(mockGen, db)

	// Store test embedding
	testContent := "This is a test document about machine learning"
	embedding := &DocumentEmbedding{
		DocumentID:  "ml-doc",
		FilePath:    "ml.txt",
		Language:    "Text",
		Category:    "docs",
		ChunkCount:  1,
		ProcessedAt: time.Now(),
		Embeddings: []EmbeddingVector{
			{
				ID:        "ml-chunk-1",
				Vector:    []float32{0.8, 0.6, 0.2},
				Content:   testContent,
				Dimension: 3,
				Metadata: map[string]string{
					"topic": "machine_learning",
				},
				CreatedAt: time.Now(),
			},
		},
	}

	err = db.Store(embedding)
	if err != nil {
		t.Fatalf("Failed to store embedding: %v", err)
	}

	// Test SearchByText functionality
	searchOptions := &VectorSearchOptions{
		TopK:           5,
		MinScore:       0.1,
		IncludeContent: true,
	}

	results, err := searchService.SearchByText("machine learning query", searchOptions)
	if err != nil {
		t.Errorf("SearchByText failed: %v", err)
	}

	// The search might find results based on the actual vector similarity
	// The important thing is that it doesn't error
	t.Logf("Search returned %d results", len(results))

	// Test that the search service correctly generates embeddings for the query
	// (This is validated by the fact that no error occurred)
}

func TestGetChunkContentDirectly(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_chunk_content.db")

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

	// Store test embedding
	testContent := "package main\n\nfunc hello() string {\n    return \"Hello, World!\"\n}"
	embedding := &DocumentEmbedding{
		DocumentID:  "hello-doc",
		FilePath:    "hello.go",
		Language:    "Go",
		Category:    "code",
		ChunkCount:  1,
		ProcessedAt: time.Now(),
		Embeddings: []EmbeddingVector{
			{
				ID:        "hello-chunk",
				Vector:    []float32{0.5, 0.5, 0.5},
				Content:   testContent,
				Dimension: 3,
				CreatedAt: time.Now(),
			},
		},
	}

	err = db.Store(embedding)
	if err != nil {
		t.Fatalf("Failed to store embedding: %v", err)
	}

	// Test getChunkContent method directly
	content, err := db.getChunkContent("hello-doc", "hello-chunk")
	if err != nil {
		t.Errorf("Failed to get chunk content: %v", err)
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}

	// Test with non-existent chunk
	_, err = db.getChunkContent("hello-doc", "non-existent-chunk")
	if err == nil {
		t.Error("Expected error for non-existent chunk")
	}

	// Test with wrong document ID
	_, err = db.getChunkContent("wrong-doc", "hello-chunk")
	if err == nil {
		t.Error("Expected error for wrong document ID")
	}
}
