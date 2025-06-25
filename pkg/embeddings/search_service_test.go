package embeddings

import (
	"testing"
)

func TestNewSearchService(t *testing.T) {
	mockGen := NewTestMockEmbeddingGenerator()
	mockDB := NewTestMockVectorDB()

	service := NewSearchService(mockGen, mockDB)

	if service == nil {
		t.Fatal("Expected non-nil search service")
	}

	if service.GetGenerator() != mockGen {
		t.Error("Generator not set correctly")
	}

	if service.GetDatabase() != mockDB {
		t.Error("Database not set correctly")
	}
}

func TestSearchByText(t *testing.T) {
	mockGen := NewTestMockEmbeddingGenerator()
	mockDB := NewTestMockVectorDB()
	service := NewSearchService(mockGen, mockDB)

	// Test successful search
	results, err := service.SearchByText("test query", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Mock DB returns empty results, which is expected
	if len(results) != 0 {
		t.Errorf("Expected 0 results from mock DB, got %d", len(results))
	}
}

func TestSearchSimilar(t *testing.T) {
	mockGen := NewTestMockEmbeddingGenerator()
	mockDB := NewTestMockVectorDB()
	service := NewSearchService(mockGen, mockDB)

	vector := []float32{1.0, 0.5, 0.3}
	results, err := service.SearchSimilar(vector, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Mock DB returns empty results
	if len(results) != 0 {
		t.Errorf("Expected 0 results from mock DB, got %d", len(results))
	}
}

func TestSearchRelated(t *testing.T) {
	mockGen := NewTestMockEmbeddingGenerator()
	mockDB := NewTestMockVectorDB()
	service := NewSearchService(mockGen, mockDB)

	// Test with non-existent document
	_, err := service.SearchRelated("non-existent", nil)
	if err == nil {
		t.Error("Expected error for non-existent document")
	}

	// Add a test document to mock DB
	testEmbedding := &DocumentEmbedding{
		DocumentID: "test-doc",
		FilePath:   "test.go",
		Embeddings: []EmbeddingVector{
			{
				ID:      "chunk-1",
				Vector:  []float32{1.0, 0.5, 0.3},
				Content: "test content",
			},
		},
	}

	err = mockDB.Store(testEmbedding)
	if err != nil {
		t.Fatalf("Failed to store test embedding: %v", err)
	}

	// Test search for related content
	results, err := service.SearchRelated("test-doc", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should return empty results since there's only one document
	// and we filter out the original document
	if len(results) != 0 {
		t.Errorf("Expected 0 results (filtered out original), got %d", len(results))
	}
}

func TestSearchRelatedWithEmptyEmbeddings(t *testing.T) {
	mockGen := NewTestMockEmbeddingGenerator()
	mockDB := NewTestMockVectorDB()
	service := NewSearchService(mockGen, mockDB)

	// Add a document with no embeddings
	testEmbedding := &DocumentEmbedding{
		DocumentID: "empty-doc",
		FilePath:   "empty.go",
		Embeddings: []EmbeddingVector{}, // No embeddings
	}

	err := mockDB.Store(testEmbedding)
	if err != nil {
		t.Fatalf("Failed to store test embedding: %v", err)
	}

	// Test search for related content
	_, err = service.SearchRelated("empty-doc", nil)
	if err == nil {
		t.Error("Expected error for document with no embeddings")
	}

	expectedError := "document empty-doc has no embeddings"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Mock implementations are already defined in embeddings_test.go
