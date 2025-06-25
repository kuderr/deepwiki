package embeddings

import (
	"fmt"
)

// SearchService provides search functionality with embedding generation
type SearchService struct {
	generator EmbeddingGenerator
	database  VectorDatabase
}

// NewSearchService creates a new search service
func NewSearchService(generator EmbeddingGenerator, database VectorDatabase) *SearchService {
	return &SearchService{
		generator: generator,
		database:  database,
	}
}

// SearchByText searches using text query by generating embeddings first
func (s *SearchService) SearchByText(text string, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	if options == nil {
		options = DefaultVectorSearchOptions()
	}

	// Generate embedding for the query text
	queryEmbedding, err := s.generator.GenerateEmbedding(text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %v", err)
	}

	// Perform vector search
	return s.database.Search(queryEmbedding, options)
}

// SearchSimilar finds similar content to the provided embedding
func (s *SearchService) SearchSimilar(embedding []float32, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	return s.database.Search(embedding, options)
}

// SearchRelated finds content related to a specific document
func (s *SearchService) SearchRelated(documentID string, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	if options == nil {
		options = DefaultVectorSearchOptions()
	}

	// Get the document first
	docEmbedding, err := s.database.Get(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document %s: %v", documentID, err)
	}

	if len(docEmbedding.Embeddings) == 0 {
		return nil, fmt.Errorf("document %s has no embeddings", documentID)
	}

	// Use the first embedding as the query vector
	// TODO: Could average multiple embeddings or use document summary
	queryVector := docEmbedding.Embeddings[0].Vector

	// Filter out the original document from results
	originalFilters := options.FilterBy
	if originalFilters == nil {
		originalFilters = make(map[string]string)
	}

	results, err := s.database.Search(queryVector, options)
	if err != nil {
		return nil, err
	}

	// Filter out chunks from the same document
	filteredResults := make([]VectorSearchResult, 0)
	for _, result := range results {
		if result.DocumentID != documentID {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, nil
}

// GetGenerator returns the embedding generator
func (s *SearchService) GetGenerator() EmbeddingGenerator {
	return s.generator
}

// GetDatabase returns the vector database
func (s *SearchService) GetDatabase() VectorDatabase {
	return s.database
}
