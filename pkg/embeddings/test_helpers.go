package embeddings

// TestMockVectorDB implements VectorDatabase for testing
type TestMockVectorDB struct {
	embeddings map[string]*DocumentEmbedding
}

// NewTestMockVectorDB creates a new mock vector database for testing
func NewTestMockVectorDB() *TestMockVectorDB {
	return &TestMockVectorDB{
		embeddings: make(map[string]*DocumentEmbedding),
	}
}

func (m *TestMockVectorDB) Store(embedding *DocumentEmbedding) error {
	m.embeddings[embedding.DocumentID] = embedding
	return nil
}

func (m *TestMockVectorDB) StoreBatch(embeddings []*DocumentEmbedding) error {
	for _, emb := range embeddings {
		if err := m.Store(emb); err != nil {
			return err
		}
	}
	return nil
}

func (m *TestMockVectorDB) Get(documentID string) (*DocumentEmbedding, error) {
	emb, exists := m.embeddings[documentID]
	if !exists {
		return nil, ErrNotFound
	}
	return emb, nil
}

func (m *TestMockVectorDB) Delete(documentID string) error {
	delete(m.embeddings, documentID)
	return nil
}

func (m *TestMockVectorDB) List() ([]string, error) {
	ids := make([]string, 0, len(m.embeddings))
	for id := range m.embeddings {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *TestMockVectorDB) Search(vector []float32, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	return []VectorSearchResult{}, nil
}

func (m *TestMockVectorDB) SearchByText(text string, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	return []VectorSearchResult{}, nil
}

func (m *TestMockVectorDB) Optimize() error {
	return nil
}

func (m *TestMockVectorDB) GetStats() (*DatabaseStats, error) {
	return &DatabaseStats{
		TotalDocuments:  len(m.embeddings),
		TotalEmbeddings: len(m.embeddings),
	}, nil
}

func (m *TestMockVectorDB) Close() error {
	return nil
}

// TestMockEmbeddingGenerator implements EmbeddingGenerator for testing
type TestMockEmbeddingGenerator struct{}

func NewTestMockEmbeddingGenerator() *TestMockEmbeddingGenerator {
	return &TestMockEmbeddingGenerator{}
}

func (m *TestMockEmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	return []float32{float32(len(text)), 0.5, 0.3}, nil
}

func (m *TestMockEmbeddingGenerator) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
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

func (m *TestMockEmbeddingGenerator) GetModel() string {
	return "mock-model"
}

func (m *TestMockEmbeddingGenerator) GetDimensions() int {
	return 3
}

func (m *TestMockEmbeddingGenerator) GetMaxTokens() int {
	return 1000
}

func (m *TestMockEmbeddingGenerator) EstimateTokens(text string) int {
	return len(text) / 4
}

func (m *TestMockEmbeddingGenerator) SplitTextForEmbedding(text string, maxTokens int) []string {
	if m.EstimateTokens(text) <= maxTokens {
		return []string{text}
	}
	mid := len(text) / 2
	return []string{text[:mid], text[mid:]}
}
