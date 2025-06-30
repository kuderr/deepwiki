package embeddings

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

// BoltVectorDB implements VectorDatabase using BoltDB for persistence
type BoltVectorDB struct {
	db         *bbolt.DB
	config     *EmbeddingConfig
	calculator *SimilarityCalculator
}

// Bucket names for different data types
const (
	documentsBucket  = "documents"
	embeddingsBucket = "embeddings"
	metadataBucket   = "metadata"
	statsBucket      = "stats"
)

// NewBoltVectorDB creates a new BoltDB-based vector database
func NewBoltVectorDB(config *EmbeddingConfig) (*BoltVectorDB, error) {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	db, err := bbolt.Open(config.StoragePath, 0o600, &bbolt.Options{
		Timeout: time.Duration(config.Timeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt database: %v", err)
	}

	// Create buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{documentsBucket, embeddingsBucket, metadataBucket, statsBucket}
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %v", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	vdb := &BoltVectorDB{
		db:         db,
		config:     config,
		calculator: NewSimilarityCalculator(CosineSimilarity),
	}

	// Initialize stats if not exist
	if err := vdb.initializeStats(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize stats: %v", err)
	}

	return vdb, nil
}

// Store stores a single document embedding
func (vdb *BoltVectorDB) Store(embedding *DocumentEmbedding) error {
	return vdb.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket([]byte(documentsBucket))
		embBucket := tx.Bucket([]byte(embeddingsBucket))

		// Store document metadata
		docData, err := json.Marshal(embedding)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %v", err)
		}

		err = docBucket.Put([]byte(embedding.DocumentID), docData)
		if err != nil {
			return fmt.Errorf("failed to store document: %v", err)
		}

		// Store individual embeddings for fast vector search
		for _, emb := range embedding.Embeddings {
			embData := &EmbeddingData{
				DocumentID: embedding.DocumentID,
				ChunkID:    emb.ID,
				FilePath:   embedding.FilePath,
				Content:    emb.Content, // Store content for easy retrieval
				Vector:     emb.Vector,
				Metadata:   emb.Metadata,
			}

			embBytes, err := json.Marshal(embData)
			if err != nil {
				return fmt.Errorf("failed to marshal embedding: %v", err)
			}

			err = embBucket.Put([]byte(emb.ID), embBytes)
			if err != nil {
				return fmt.Errorf("failed to store embedding: %v", err)
			}
		}

		// Update stats
		return vdb.updateStatsInTx(tx, 1, len(embedding.Embeddings))
	})
}

// StoreBatch stores multiple document embeddings in a single transaction
func (vdb *BoltVectorDB) StoreBatch(embeddings []*DocumentEmbedding) error {
	return vdb.db.Update(func(tx *bbolt.Tx) error {
		docBucket := tx.Bucket([]byte(documentsBucket))
		embBucket := tx.Bucket([]byte(embeddingsBucket))

		totalEmbeddings := 0

		for _, embedding := range embeddings {
			// Store document metadata
			docData, err := json.Marshal(embedding)
			if err != nil {
				return fmt.Errorf("failed to marshal document %s: %v", embedding.DocumentID, err)
			}

			err = docBucket.Put([]byte(embedding.DocumentID), docData)
			if err != nil {
				return fmt.Errorf("failed to store document %s: %v", embedding.DocumentID, err)
			}

			// Store individual embeddings
			for _, emb := range embedding.Embeddings {
				embData := &EmbeddingData{
					DocumentID: embedding.DocumentID,
					ChunkID:    emb.ID,
					FilePath:   embedding.FilePath,
					Content:    emb.Content, // Store content for easy retrieval
					Vector:     emb.Vector,
					Metadata:   emb.Metadata,
				}

				embBytes, err := json.Marshal(embData)
				if err != nil {
					return fmt.Errorf("failed to marshal embedding %s: %v", emb.ID, err)
				}

				err = embBucket.Put([]byte(emb.ID), embBytes)
				if err != nil {
					return fmt.Errorf("failed to store embedding %s: %v", emb.ID, err)
				}
			}

			totalEmbeddings += len(embedding.Embeddings)
		}

		// Update stats
		return vdb.updateStatsInTx(tx, len(embeddings), totalEmbeddings)
	})
}

// Get retrieves a document embedding by ID
func (vdb *BoltVectorDB) Get(documentID string) (*DocumentEmbedding, error) {
	var embedding *DocumentEmbedding

	err := vdb.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(documentsBucket))
		data := bucket.Get([]byte(documentID))
		if data == nil {
			return fmt.Errorf("document not found: %s", documentID)
		}

		return json.Unmarshal(data, &embedding)
	})

	return embedding, err
}

// Delete removes a document and all its embeddings
func (vdb *BoltVectorDB) Delete(documentID string) error {
	return vdb.db.Update(func(tx *bbolt.Tx) error {
		// Get document first to find all embedding IDs
		docBucket := tx.Bucket([]byte(documentsBucket))
		embBucket := tx.Bucket([]byte(embeddingsBucket))

		data := docBucket.Get([]byte(documentID))
		if data == nil {
			return fmt.Errorf("document not found: %s", documentID)
		}

		var embedding DocumentEmbedding
		err := json.Unmarshal(data, &embedding)
		if err != nil {
			return fmt.Errorf("failed to unmarshal document: %v", err)
		}

		// Delete individual embeddings
		for _, emb := range embedding.Embeddings {
			err = embBucket.Delete([]byte(emb.ID))
			if err != nil {
				return fmt.Errorf("failed to delete embedding %s: %v", emb.ID, err)
			}
		}

		// Delete document
		err = docBucket.Delete([]byte(documentID))
		if err != nil {
			return fmt.Errorf("failed to delete document: %v", err)
		}

		// Update stats
		return vdb.updateStatsInTx(tx, -1, -len(embedding.Embeddings))
	})
}

// List returns all document IDs
func (vdb *BoltVectorDB) List() ([]string, error) {
	var documentIDs []string

	err := vdb.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(documentsBucket))
		return bucket.ForEach(func(k, v []byte) error {
			documentIDs = append(documentIDs, string(k))
			return nil
		})
	})

	return documentIDs, err
}

// Search performs vector similarity search
func (vdb *BoltVectorDB) Search(vector []float32, options *VectorSearchOptions) ([]VectorSearchResult, error) {
	if options == nil {
		options = DefaultVectorSearchOptions()
	}

	var results []VectorSearchResult
	candidates := make([]candidateResult, 0)

	err := vdb.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(embeddingsBucket))

		// Iterate through all embeddings and calculate similarity
		return bucket.ForEach(func(k, v []byte) error {
			var embData EmbeddingData
			err := json.Unmarshal(v, &embData)
			if err != nil {
				return nil // Skip invalid entries
			}

			// Apply filters
			if !vdb.matchesFilters(&embData, options.FilterBy) {
				return nil
			}

			// Calculate similarity
			score := vdb.calculator.Calculate(vector, embData.Vector)
			if score < options.MinScore {
				return nil
			}

			candidates = append(candidates, candidateResult{
				DocumentID: embData.DocumentID,
				ChunkID:    embData.ChunkID,
				FilePath:   embData.FilePath,
				Score:      score,
				Metadata:   embData.Metadata,
			})

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	// Sort by score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Apply limits
	limit := options.TopK
	if limit > len(candidates) {
		limit = len(candidates)
	}

	// Convert to results
	results = make([]VectorSearchResult, limit)
	for i := 0; i < limit; i++ {
		candidate := candidates[i]
		results[i] = VectorSearchResult{
			DocumentID: candidate.DocumentID,
			ChunkID:    candidate.ChunkID,
			FilePath:   candidate.FilePath,
			Score:      candidate.Score,
			Metadata:   candidate.Metadata,
		}

		// Add content if requested
		if options.IncludeContent {
			// TODO: Optimize this by storing content in embedding data or using a single query
			content, err := vdb.getChunkContent(candidate.DocumentID, candidate.ChunkID)
			if err == nil {
				results[i].Content = content
			} else {
				// Fallback to empty content if lookup fails
				results[i].Content = ""
			}
		}
	}

	return results, nil
}

// Optimize optimizes the database (compact, rebuild indexes, etc.)
func (vdb *BoltVectorDB) Optimize() error {
	// BoltDB doesn't require explicit optimization, but we can update stats
	return vdb.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(statsBucket))
		stats := &DatabaseStats{}

		data := bucket.Get([]byte("stats"))
		if data != nil {
			json.Unmarshal(data, stats)
		}

		stats.LastOptimized = time.Now()

		statsBytes, err := json.Marshal(stats)
		if err != nil {
			return err
		}

		return bucket.Put([]byte("stats"), statsBytes)
	})
}

// GetStats returns database statistics
func (vdb *BoltVectorDB) GetStats() (*DatabaseStats, error) {
	var stats *DatabaseStats

	err := vdb.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(statsBucket))
		data := bucket.Get([]byte("stats"))
		if data == nil {
			stats = &DatabaseStats{}
			return nil
		}

		return json.Unmarshal(data, &stats)
	})

	return stats, err
}

// Close closes the database
func (vdb *BoltVectorDB) Close() error {
	return vdb.db.Close()
}

// getChunkContent retrieves the content for a specific chunk
func (vdb *BoltVectorDB) getChunkContent(documentID, chunkID string) (string, error) {
	var content string

	err := vdb.db.View(func(tx *bbolt.Tx) error {
		embBucket := tx.Bucket([]byte(embeddingsBucket))
		embData := embBucket.Get([]byte(chunkID))
		if embData == nil {
			return fmt.Errorf("chunk not found: %s", chunkID)
		}

		var embedding EmbeddingData
		err := json.Unmarshal(embData, &embedding)
		if err != nil {
			return fmt.Errorf("failed to unmarshal embedding: %v", err)
		}

		// Verify the document ID matches
		if embedding.DocumentID != documentID {
			return fmt.Errorf("chunk %s does not belong to document %s", chunkID, documentID)
		}

		content = embedding.Content
		return nil
	})

	return content, err
}

// Helper types and functions

// EmbeddingData represents stored embedding data
type EmbeddingData struct {
	DocumentID string            `json:"documentId"`
	ChunkID    string            `json:"chunkId"`
	FilePath   string            `json:"filePath"`
	Content    string            `json:"content"` // Store chunk content for easy retrieval
	Vector     []float32         `json:"vector"`
	Metadata   map[string]string `json:"metadata"`
}

// candidateResult represents a search candidate
type candidateResult struct {
	DocumentID string
	ChunkID    string
	FilePath   string
	Score      float32
	Metadata   map[string]string
}

// initializeStats initializes database statistics
func (vdb *BoltVectorDB) initializeStats() error {
	return vdb.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(statsBucket))
		data := bucket.Get([]byte("stats"))
		if data != nil {
			return nil // Stats already exist
		}

		stats := &DatabaseStats{
			TotalDocuments:   0,
			TotalEmbeddings:  0,
			TotalSize:        0,
			IndexSize:        0,
			LastOptimized:    time.Now(),
			AverageDimension: vdb.config.Dimensions,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		statsBytes, err := json.Marshal(stats)
		if err != nil {
			return err
		}

		return bucket.Put([]byte("stats"), statsBytes)
	})
}

// updateStatsInTx updates statistics within a transaction
func (vdb *BoltVectorDB) updateStatsInTx(tx *bbolt.Tx, docDelta, embDelta int) error {
	bucket := tx.Bucket([]byte(statsBucket))

	var stats DatabaseStats
	data := bucket.Get([]byte("stats"))
	if data != nil {
		err := json.Unmarshal(data, &stats)
		if err != nil {
			return err
		}
	}

	stats.TotalDocuments += docDelta
	stats.TotalEmbeddings += embDelta
	stats.UpdatedAt = time.Now()

	// Calculate database size (approximation)
	stats.TotalSize = int64(stats.TotalEmbeddings * stats.AverageDimension * 4) // 4 bytes per float32

	statsBytes, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	return bucket.Put([]byte("stats"), statsBytes)
}

// matchesFilters checks if embedding data matches the provided filters
func (vdb *BoltVectorDB) matchesFilters(embData *EmbeddingData, filters map[string]string) bool {
	for key, value := range filters {
		if embData.Metadata[key] != value {
			return false
		}
	}
	return true
}
