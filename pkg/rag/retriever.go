package rag

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/embeddings"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/processor"
)

// DefaultDocumentRetriever implements DocumentRetriever using embeddings and text processing
type DefaultDocumentRetriever struct {
	embeddingService *embeddings.EmbeddingService
	vectorDB         embeddings.VectorDatabase
	embeddingGen     embeddings.EmbeddingGenerator
	documents        []processor.Document
	config           *RAGConfig
	cache            RetrievalCache
	stats            *RetrievalStats
	mu               sync.RWMutex
}

// NewDocumentRetriever creates a new document retriever
func NewDocumentRetriever(
	embeddingService *embeddings.EmbeddingService,
	vectorDB embeddings.VectorDatabase,
	embeddingGen embeddings.EmbeddingGenerator,
	documents []processor.Document,
	config *RAGConfig,
) *DefaultDocumentRetriever {
	if config == nil {
		config = DefaultRAGConfig()
	}

	retriever := &DefaultDocumentRetriever{
		embeddingService: embeddingService,
		vectorDB:         vectorDB,
		embeddingGen:     embeddingGen,
		documents:        documents,
		config:           config,
		stats: &RetrievalStats{
			PerformanceMetrics: make(map[string]float64),
			MostCommonQueries:  make([]QueryStats, 0),
		},
	}

	// Initialize cache if enabled
	if config.CacheResults {
		retriever.cache = NewInMemoryCache(1000) // Default cache size
	}

	return retriever
}

// RetrieveRelevantDocuments retrieves documents based on a retrieval context
func (r *DefaultDocumentRetriever) RetrieveRelevantDocuments(ctx *RetrievalContext) ([]RetrievalResult, error) {
	startTime := time.Now()
	defer func() {
		r.updateStats("RetrieveRelevantDocuments", time.Since(startTime))
	}()

	// Check cache first
	if r.cache != nil {
		cacheKey := r.generateCacheKey(ctx)
		if cached, found := r.cache.Get(cacheKey); found {
			r.stats.CacheHits++
			return cached, nil
		}
		r.stats.CacheMisses++
	}

	var results []RetrievalResult
	var err error

	// Route to appropriate retrieval strategy
	switch ctx.QueryType {
	case QueryTypeSemantic:
		results, err = r.retrieveSemantic(ctx)
	case QueryTypeKeyword:
		results, err = r.retrieveKeyword(ctx)
	case QueryTypeHybrid:
		results, err = r.retrieveHybrid(ctx)
	case QueryTypeStructural:
		results, err = r.retrieveStructural(ctx)
	default:
		results, err = r.retrieveHybrid(ctx) // Default to hybrid
	}

	if err != nil {
		return nil, err
	}

	// Apply filters
	results = r.FilterResults(results, ctx.Filters)

	// Apply time window filter if specified
	if ctx.TimeWindow != nil {
		results = r.filterByTimeWindow(results, ctx.TimeWindow)
	}

	// Rerank if enabled
	if r.config.RerankResults {
		results, err = r.RerankResults(results, ctx.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to rerank results: %v", err)
		}
	}

	// Apply diversity filtering
	results = r.applyDiversityFiltering(results)

	// Limit results
	if ctx.MaxResults > 0 && len(results) > ctx.MaxResults {
		results = results[:ctx.MaxResults]
	}

	// Add context if enabled
	if r.config.IncludeContext {
		results = r.enrichWithContext(results)
	}

	// Cache results
	if r.cache != nil {
		cacheKey := r.generateCacheKey(ctx)
		r.cache.Set(cacheKey, results, r.config.CacheTTL)
	}

	return results, nil
}

// RetrieveByQuery retrieves documents using a simple query
func (r *DefaultDocumentRetriever) RetrieveByQuery(query string, maxResults int) ([]RetrievalResult, error) {
	ctx := &RetrievalContext{
		Query:      query,
		QueryType:  r.config.RetrievalStrategy,
		MaxResults: maxResults,
		MinScore:   r.config.DefaultMinScore,
		Filters:    make(map[string]string),
	}

	return r.RetrieveRelevantDocuments(ctx)
}

// RetrieveByTags retrieves documents by tags/metadata
func (r *DefaultDocumentRetriever) RetrieveByTags(tags []string, maxResults int) ([]RetrievalResult, error) {
	results := make([]RetrievalResult, 0)

	// TODO: Implement more sophisticated tag matching with fuzzy search and synonyms
	for _, doc := range r.documents {
		for _, chunk := range doc.Chunks {
			matches := 0
			for _, tag := range tags {
				// Check if tag appears in content or metadata
				if strings.Contains(strings.ToLower(chunk.Text), strings.ToLower(tag)) {
					matches++
				}
				for _, value := range chunk.Metadata {
					if strings.Contains(strings.ToLower(value), strings.ToLower(tag)) {
						matches++
					}
				}
			}

			if matches > 0 {
				result := RetrievalResult{
					DocumentID: doc.ID,
					ChunkID:    chunk.ID,
					FilePath:   doc.FilePath,
					Content:    chunk.Text,
					Score:      float32(matches) / float32(len(tags)),
					Language:   doc.Language,
					Category:   doc.Category,
					Metadata:   chunk.Metadata,
					Relevance: RelevanceInfo{
						RelevanceScore: float32(matches) / float32(len(tags)),
						MatchedTerms:   tags[:matches],
					},
				}
				results = append(results, result)
			}
		}

		if len(results) >= maxResults {
			break
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// RetrieveCodeExamples retrieves code examples for a specific language and concept
func (r *DefaultDocumentRetriever) RetrieveCodeExamples(language string, concept string, maxResults int) ([]RetrievalResult, error) {
	ctx := &RetrievalContext{
		Query:      concept,
		QueryType:  QueryTypeHybrid,
		MaxResults: maxResults * 2, // Get more to filter
		MinScore:   r.config.DefaultMinScore,
		Filters: map[string]string{
			"language": language,
			"category": "code",
		},
	}

	results, err := r.RetrieveRelevantDocuments(ctx)
	if err != nil {
		return nil, err
	}

	// Filter for code examples
	codeResults := make([]RetrievalResult, 0)
	for _, result := range results {
		if result.Language == language && result.Category == "code" {
			// Boost score for function/class definitions
			if r.isCodeDefinition(result.Content) {
				result.Score *= 1.2
			}
			codeResults = append(codeResults, result)
		}
	}

	// Sort by score and limit
	sort.Slice(codeResults, func(i, j int) bool {
		return codeResults[i].Score > codeResults[j].Score
	})

	if len(codeResults) > maxResults {
		codeResults = codeResults[:maxResults]
	}

	return codeResults, nil
}

// RetrieveDocumentation retrieves documentation content
func (r *DefaultDocumentRetriever) RetrieveDocumentation(query string, maxResults int) ([]RetrievalResult, error) {
	ctx := &RetrievalContext{
		Query:      query,
		QueryType:  QueryTypeHybrid,
		MaxResults: maxResults,
		MinScore:   r.config.DefaultMinScore,
		Filters: map[string]string{
			"category": "docs",
		},
	}

	return r.RetrieveRelevantDocuments(ctx)
}

// RetrieveConfigFiles retrieves configuration files
func (r *DefaultDocumentRetriever) RetrieveConfigFiles(configType string, maxResults int) ([]RetrievalResult, error) {
	ctx := &RetrievalContext{
		Query:      configType,
		QueryType:  QueryTypeKeyword,
		MaxResults: maxResults,
		MinScore:   0.1,
		Filters: map[string]string{
			"category": "config",
		},
	}

	return r.RetrieveRelevantDocuments(ctx)
}

// RetrieveWithContext retrieves documents considering existing context
func (r *DefaultDocumentRetriever) RetrieveWithContext(query string, context []RetrievalResult, maxResults int) ([]RetrievalResult, error) {
	// Extract context terms and boost related content
	contextTerms := r.extractContextTerms(context)

	ctx := &RetrievalContext{
		Query:      query,
		QueryType:  QueryTypeHybrid,
		MaxResults: maxResults,
		MinScore:   r.config.DefaultMinScore,
		BoostFactors: map[string]float32{
			"context_match": 1.3,
		},
	}

	results, err := r.RetrieveRelevantDocuments(ctx)
	if err != nil {
		return nil, err
	}

	// Boost results that match context terms
	for i := range results {
		for _, term := range contextTerms {
			if strings.Contains(strings.ToLower(results[i].Content), strings.ToLower(term)) {
				results[i].Score *= 1.1
			}
		}
	}

	// Re-sort after boosting
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// RetrieveRelatedChunks finds chunks related to a given chunk
func (r *DefaultDocumentRetriever) RetrieveRelatedChunks(chunkID string, maxResults int) ([]RetrievalResult, error) {
	// Find the source chunk
	var sourceChunk *processor.TextChunk

	for i := range r.documents {
		for j := range r.documents[i].Chunks {
			if r.documents[i].Chunks[j].ID == chunkID {
				sourceChunk = &r.documents[i].Chunks[j]
				break
			}
		}
		if sourceChunk != nil {
			break
		}
	}

	if sourceChunk == nil {
		return nil, fmt.Errorf("chunk not found: %s", chunkID)
	}

	// Use the chunk content as query
	ctx := &RetrievalContext{
		Query:      sourceChunk.Text,
		QueryType:  QueryTypeSemantic,
		MaxResults: maxResults + 1, // +1 because we'll exclude the original
		MinScore:   0.3,            // Higher threshold for related content
	}

	results, err := r.RetrieveRelevantDocuments(ctx)
	if err != nil {
		return nil, err
	}

	// Filter out the original chunk
	filteredResults := make([]RetrievalResult, 0)
	for _, result := range results {
		if result.ChunkID != chunkID {
			filteredResults = append(filteredResults, result)
		}
	}

	if len(filteredResults) > maxResults {
		filteredResults = filteredResults[:maxResults]
	}

	return filteredResults, nil
}

// FilterResults filters results based on metadata filters
func (r *DefaultDocumentRetriever) FilterResults(results []RetrievalResult, filters map[string]string) []RetrievalResult {
	if len(filters) == 0 {
		return results
	}

	filtered := make([]RetrievalResult, 0)
	for _, result := range results {
		matches := true

		for key, value := range filters {
			switch key {
			case "language":
				if result.Language != value {
					matches = false
				}
			case "category":
				if result.Category != value {
					matches = false
				}
			case "filePath":
				if !strings.Contains(result.FilePath, value) {
					matches = false
				}
			default:
				if result.Metadata[key] != value {
					matches = false
				}
			}

			if !matches {
				break
			}
		}

		if matches {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// RerankResults reranks results based on the query
func (r *DefaultDocumentRetriever) RerankResults(results []RetrievalResult, query string) ([]RetrievalResult, error) {
	// Simple reranking based on keyword matching and other factors
	queryTerms := strings.Fields(strings.ToLower(query))

	for i := range results {
		// Calculate new relevance score
		keywordScore := r.calculateKeywordScore(results[i].Content, queryTerms)
		structuralScore := r.calculateStructuralScore(results[i])

		// Combine scores with weights
		newScore := results[i].Score*r.config.SemanticWeight +
			keywordScore*r.config.KeywordWeight +
			structuralScore*r.config.StructuralWeight

		results[i].Relevance = RelevanceInfo{
			RelevanceScore:  newScore,
			SemanticScore:   results[i].Score,
			KeywordScore:    keywordScore,
			StructuralScore: structuralScore,
			MatchedTerms:    r.findMatchedTerms(results[i].Content, queryTerms),
		}

		results[i].Score = newScore
	}

	// Sort by new score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// GetRetrievalStats returns retrieval statistics
func (r *DefaultDocumentRetriever) GetRetrievalStats() *RetrievalStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to return
	stats := *r.stats
	return &stats
}

// ClearCache clears the retrieval cache
func (r *DefaultDocumentRetriever) ClearCache() error {
	if r.cache != nil {
		r.cache.Clear()
	}
	return nil
}

// Helper methods

// retrieveSemantic performs semantic search using embeddings
func (r *DefaultDocumentRetriever) retrieveSemantic(ctx *RetrievalContext) ([]RetrievalResult, error) {
	// Generate query embedding
	queryEmbedding, err := r.embeddingGen.GenerateEmbedding(ctx.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %v", err)
	}

	// Search vector database
	searchOptions := &embeddings.VectorSearchOptions{
		TopK:           ctx.MaxResults,
		MinScore:       ctx.MinScore,
		FilterBy:       ctx.Filters,
		IncludeContent: true,
	}

	vectorResults, err := r.vectorDB.Search(queryEmbedding, searchOptions)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %v", err)
	}

	// Convert to retrieval results
	results := make([]RetrievalResult, len(vectorResults))
	for i, vr := range vectorResults {
		results[i] = RetrievalResult{
			DocumentID: vr.DocumentID,
			ChunkID:    vr.ChunkID,
			FilePath:   vr.FilePath,
			Content:    vr.Content,
			Score:      vr.Score,
			Metadata:   vr.Metadata,
			Relevance: RelevanceInfo{
				SemanticScore: vr.Score,
			},
		}

		// Enrich with document info
		r.enrichWithDocumentInfo(&results[i])
	}

	return results, nil
}

// retrieveKeyword performs keyword-based search
func (r *DefaultDocumentRetriever) retrieveKeyword(ctx *RetrievalContext) ([]RetrievalResult, error) {
	queryTerms := strings.Fields(strings.ToLower(ctx.Query))
	results := make([]RetrievalResult, 0)

	for _, doc := range r.documents {
		for _, chunk := range doc.Chunks {
			score := r.calculateKeywordScore(chunk.Text, queryTerms)
			if score >= ctx.MinScore {
				result := RetrievalResult{
					DocumentID: doc.ID,
					ChunkID:    chunk.ID,
					FilePath:   doc.FilePath,
					Content:    chunk.Text,
					Score:      score,
					Language:   doc.Language,
					Category:   doc.Category,
					Metadata:   chunk.Metadata,
					Relevance: RelevanceInfo{
						KeywordScore: score,
						MatchedTerms: r.findMatchedTerms(chunk.Text, queryTerms),
					},
				}
				results = append(results, result)
			}
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// retrieveHybrid combines semantic and keyword search
func (r *DefaultDocumentRetriever) retrieveHybrid(ctx *RetrievalContext) ([]RetrievalResult, error) {
	// Get semantic results
	semanticResults, err := r.retrieveSemantic(ctx)
	if err != nil {
		return nil, err
	}

	// Get keyword results
	keywordResults, err := r.retrieveKeyword(ctx)
	if err != nil {
		return nil, err
	}

	// Combine and deduplicate results
	resultMap := make(map[string]*RetrievalResult)

	// Add semantic results
	for _, result := range semanticResults {
		resultMap[result.ChunkID] = &result
	}

	// Merge keyword results
	for _, kwResult := range keywordResults {
		if existing, found := resultMap[kwResult.ChunkID]; found {
			// Combine scores
			existing.Score = existing.Score*r.config.SemanticWeight + kwResult.Score*r.config.KeywordWeight
			existing.Relevance.KeywordScore = kwResult.Score
			existing.Relevance.MatchedTerms = kwResult.Relevance.MatchedTerms
		} else {
			kwResult.Score = kwResult.Score * r.config.KeywordWeight
			resultMap[kwResult.ChunkID] = &kwResult
		}
	}

	// Convert back to slice
	results := make([]RetrievalResult, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, *result)
	}

	// Sort by combined score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// retrieveStructural performs structure-based search (functions, classes, etc.)
func (r *DefaultDocumentRetriever) retrieveStructural(ctx *RetrievalContext) ([]RetrievalResult, error) {
	results := make([]RetrievalResult, 0)
	queryLower := strings.ToLower(ctx.Query)

	for _, doc := range r.documents {
		if doc.Category != "code" {
			continue // Only apply to code files
		}

		for _, chunk := range doc.Chunks {
			score := r.calculateStructuralScore(RetrievalResult{
				Content:  chunk.Text,
				Language: doc.Language,
			})

			// Check if query matches structural elements
			if r.matchesStructuralQuery(chunk.Text, queryLower, doc.Language) {
				score *= 2.0 // Boost structural matches
			}

			if score >= ctx.MinScore {
				result := RetrievalResult{
					DocumentID: doc.ID,
					ChunkID:    chunk.ID,
					FilePath:   doc.FilePath,
					Content:    chunk.Text,
					Score:      score,
					Language:   doc.Language,
					Category:   doc.Category,
					Metadata:   chunk.Metadata,
					Relevance: RelevanceInfo{
						StructuralScore: score,
					},
				}
				results = append(results, result)
			}
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// Helper functions for scoring and matching

func (r *DefaultDocumentRetriever) calculateKeywordScore(content string, queryTerms []string) float32 {
	contentLower := strings.ToLower(content)
	matches := 0
	totalTerms := len(queryTerms)

	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			matches++
		}
	}

	if totalTerms == 0 {
		return 0.0
	}

	return float32(matches) / float32(totalTerms)
}

func (r *DefaultDocumentRetriever) calculateStructuralScore(result RetrievalResult) float32 {
	score := float32(0.1) // Base score

	content := result.Content

	// Boost for function definitions
	if r.isFunctionDefinition(content, result.Language) {
		score += 0.3
	}

	// Boost for class definitions
	if r.isClassDefinition(content, result.Language) {
		score += 0.3
	}

	// Boost for interface definitions
	if r.isInterfaceDefinition(content, result.Language) {
		score += 0.2
	}

	// Boost for important keywords
	if r.containsImportantKeywords(content, result.Language) {
		score += 0.1
	}

	return score
}

func (r *DefaultDocumentRetriever) findMatchedTerms(content string, queryTerms []string) []string {
	contentLower := strings.ToLower(content)
	matched := make([]string, 0)

	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			matched = append(matched, term)
		}
	}

	return matched
}

func (r *DefaultDocumentRetriever) isCodeDefinition(content string) bool {
	// TODO: Use the LanguageSpecificProcessor`s from pkg/processor for more accurate code structure detection
	definitionKeywords := []string{"func ", "function ", "class ", "def ", "interface ", "type "}
	contentLower := strings.ToLower(content)

	for _, keyword := range definitionKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}

	return false
}

func (r *DefaultDocumentRetriever) isFunctionDefinition(content, language string) bool {
	contentLower := strings.ToLower(content)

	switch language {
	case "Go":
		return strings.Contains(contentLower, "func ")
	case "Python":
		return strings.Contains(contentLower, "def ")
	case "JavaScript", "TypeScript":
		return strings.Contains(contentLower, "function ") || strings.Contains(contentLower, "const ") || strings.Contains(contentLower, "let ")
	case "Java", "C#":
		return strings.Contains(contentLower, "public ") || strings.Contains(contentLower, "private ") || strings.Contains(contentLower, "protected ")
	default:
		return strings.Contains(contentLower, "function") || strings.Contains(contentLower, "def") || strings.Contains(contentLower, "func")
	}
}

func (r *DefaultDocumentRetriever) isClassDefinition(content, language string) bool {
	return strings.Contains(strings.ToLower(content), "class ")
}

func (r *DefaultDocumentRetriever) isInterfaceDefinition(content, language string) bool {
	return strings.Contains(strings.ToLower(content), "interface ")
}

func (r *DefaultDocumentRetriever) containsImportantKeywords(content, language string) bool {
	importantKeywords := []string{"import", "export", "main", "init", "setup", "config"}
	contentLower := strings.ToLower(content)

	for _, keyword := range importantKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}

	return false
}

func (r *DefaultDocumentRetriever) matchesStructuralQuery(content, query, language string) bool {
	// Check if query matches function/class names or structural elements
	return strings.Contains(strings.ToLower(content), query)
}

// Additional helper functions

func (r *DefaultDocumentRetriever) generateCacheKey(ctx *RetrievalContext) string {
	// TODO: Use hash-based cache keys and include all relevant context parameters
	return fmt.Sprintf("%s_%s_%d_%f", ctx.Query, ctx.QueryType, ctx.MaxResults, ctx.MinScore)
}

func (r *DefaultDocumentRetriever) enrichWithDocumentInfo(result *RetrievalResult) {
	// Find the source document and enrich the result
	for _, doc := range r.documents {
		if doc.ID == result.DocumentID {
			result.Language = doc.Language
			result.Category = doc.Category
			break
		}
	}
}

func (r *DefaultDocumentRetriever) enrichWithContext(results []RetrievalResult) []RetrievalResult {
	// TODO: Implement proper context enrichment with surrounding chunks and code structure
	for i := range results {
		// This would involve finding surrounding chunks, function context, etc.
		results[i].Context = &ChunkContext{
			// Populate with actual context information
		}
	}
	return results
}

func (r *DefaultDocumentRetriever) extractContextTerms(context []RetrievalResult) []string {
	terms := make(map[string]bool)

	for _, result := range context {
		words := strings.Fields(strings.ToLower(result.Content))
		for _, word := range words {
			if len(word) > 3 { // Filter short words
				terms[word] = true
			}
		}
	}

	contextTerms := make([]string, 0, len(terms))
	for term := range terms {
		contextTerms = append(contextTerms, term)
	}

	return contextTerms
}

func (r *DefaultDocumentRetriever) filterByTimeWindow(results []RetrievalResult, timeWindow *TimeWindow) []RetrievalResult {
	// Simple time filtering - would need to get file modification times
	// For now, return all results
	return results
}

func (r *DefaultDocumentRetriever) applyDiversityFiltering(results []RetrievalResult) []RetrievalResult {
	// Simple diversity filtering to avoid too many similar results
	if len(results) <= r.config.MaxSimilarResults {
		return results
	}

	diverse := make([]RetrievalResult, 0)
	filePathSeen := make(map[string]int)

	for _, result := range results {
		count := filePathSeen[result.FilePath]
		if count < r.config.MaxSimilarResults {
			diverse = append(diverse, result)
			filePathSeen[result.FilePath] = count + 1
		}
	}

	return diverse
}

func (r *DefaultDocumentRetriever) updateStats(operation string, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats.TotalQueries++

	// Update average query time
	if r.stats.AverageQueryTime == 0 {
		r.stats.AverageQueryTime = duration
	} else {
		r.stats.AverageQueryTime = (r.stats.AverageQueryTime + duration) / 2
	}

	// Update performance metrics
	r.stats.PerformanceMetrics[operation] = float64(duration.Milliseconds())
}
