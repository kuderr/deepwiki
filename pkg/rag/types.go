package rag

import (
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/processor"
)

// RetrievalContext represents the context for document retrieval
type RetrievalContext struct {
	Query        string             `json:"query"`        // The search query
	QueryType    QueryType          `json:"queryType"`    // Type of query (semantic, keyword, hybrid)
	MaxResults   int                `json:"maxResults"`   // Maximum number of results to return
	MinScore     float32            `json:"minScore"`     // Minimum similarity score
	Filters      map[string]string  `json:"filters"`      // Metadata filters
	BoostFactors map[string]float32 `json:"boostFactors"` // Boost factors for different attributes
	TimeWindow   *TimeWindow        `json:"timeWindow"`   // Optional time window for filtering
}

// QueryType represents different types of queries
type QueryType string

const (
	QueryTypeSemantic   QueryType = "semantic"   // Vector-based semantic search
	QueryTypeKeyword    QueryType = "keyword"    // Keyword-based search
	QueryTypeHybrid     QueryType = "hybrid"     // Combination of semantic and keyword
	QueryTypeStructural QueryType = "structural" // Structure-based search (classes, functions, etc.)
)

// TimeWindow represents a time range for filtering documents
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RetrievalResult represents a retrieved document chunk with context
type RetrievalResult struct {
	DocumentID string            `json:"documentId"` // Source document ID
	ChunkID    string            `json:"chunkId"`    // Source chunk ID
	FilePath   string            `json:"filePath"`   // Original file path
	Content    string            `json:"content"`    // Chunk content
	Score      float32           `json:"score"`      // Relevance score
	Language   string            `json:"language"`   // Programming language
	Category   string            `json:"category"`   // File category
	Context    *ChunkContext     `json:"context"`    // Surrounding context
	Metadata   map[string]string `json:"metadata"`   // Additional metadata
	Relevance  RelevanceInfo     `json:"relevance"`  // Relevance information
}

// ChunkContext provides context around a retrieved chunk
type ChunkContext struct {
	PreviousChunk *processor.TextChunk `json:"previousChunk"` // Previous chunk in document
	NextChunk     *processor.TextChunk `json:"nextChunk"`     // Next chunk in document
	DocumentStart string               `json:"documentStart"` // Beginning of document
	FunctionName  string               `json:"functionName"`  // Function/method name if applicable
	ClassName     string               `json:"className"`     // Class name if applicable
	LineNumbers   []int                `json:"lineNumbers"`   // Source line numbers
}

// RelevanceInfo provides information about why a chunk is relevant
type RelevanceInfo struct {
	RelevanceScore  float32  `json:"relevanceScore"`  // Overall relevance score
	SemanticScore   float32  `json:"semanticScore"`   // Semantic similarity score
	KeywordScore    float32  `json:"keywordScore"`    // Keyword match score
	StructuralScore float32  `json:"structuralScore"` // Structural relevance score
	MatchedTerms    []string `json:"matchedTerms"`    // Keywords/terms that matched
	MatchedConcepts []string `json:"matchedConcepts"` // Semantic concepts that matched
	BoostFactors    []string `json:"boostFactors"`    // Applied boost factors
}

// RAGConfig represents configuration for the RAG system
type RAGConfig struct {
	// Retrieval settings
	DefaultMaxResults int       `json:"defaultMaxResults"` // Default max results
	DefaultMinScore   float32   `json:"defaultMinScore"`   // Default min score
	RetrievalStrategy QueryType `json:"retrievalStrategy"` // Default retrieval strategy

	// Context settings
	ContextWindow  int  `json:"contextWindow"`  // Context window size (chunks)
	IncludeContext bool `json:"includeContext"` // Whether to include context
	ContextOverlap int  `json:"contextOverlap"` // Context overlap size

	// Ranking settings
	RerankResults    bool    `json:"rerankResults"`    // Whether to rerank results
	RerankingModel   string  `json:"rerankingModel"`   // Model for reranking
	SemanticWeight   float32 `json:"semanticWeight"`   // Weight for semantic score
	KeywordWeight    float32 `json:"keywordWeight"`    // Weight for keyword score
	StructuralWeight float32 `json:"structuralWeight"` // Weight for structural score

	// Filtering settings
	FilterByLanguage   bool `json:"filterByLanguage"`   // Filter by programming language
	FilterByCategory   bool `json:"filterByCategory"`   // Filter by file category
	FilterByImportance bool `json:"filterByImportance"` // Filter by importance score
	MinImportance      int  `json:"minImportance"`      // Minimum importance score

	// Diversity settings
	DiversityThreshold float32 `json:"diversityThreshold"` // Threshold for diversity filtering
	MaxSimilarResults  int     `json:"maxSimilarResults"`  // Max similar results to include

	// Performance settings
	CacheResults      bool          `json:"cacheResults"`      // Whether to cache results
	CacheTTL          time.Duration `json:"cacheTTL"`          // Cache time-to-live
	ConcurrentQueries int           `json:"concurrentQueries"` // Max concurrent queries
}

// DefaultRAGConfig returns default RAG configuration
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
		DefaultMaxResults:  20,
		DefaultMinScore:    0.1,
		RetrievalStrategy:  QueryTypeHybrid,
		ContextWindow:      2,
		IncludeContext:     true,
		ContextOverlap:     1,
		RerankResults:      true,
		RerankingModel:     "",
		SemanticWeight:     0.6,
		KeywordWeight:      0.3,
		StructuralWeight:   0.1,
		FilterByLanguage:   false,
		FilterByCategory:   false,
		FilterByImportance: false,
		MinImportance:      1,
		DiversityThreshold: 0.9,
		MaxSimilarResults:  3,
		CacheResults:       true,
		CacheTTL:           30 * time.Minute,
		ConcurrentQueries:  5,
	}
}

// DocumentRetriever interface for retrieving relevant documents
type DocumentRetriever interface {
	// Primary retrieval methods
	RetrieveRelevantDocuments(ctx *RetrievalContext) ([]RetrievalResult, error)
	RetrieveByQuery(query string, maxResults int) ([]RetrievalResult, error)
	RetrieveByTags(tags []string, maxResults int) ([]RetrievalResult, error)

	// Specialized retrieval methods
	RetrieveCodeExamples(language, concept string, maxResults int) ([]RetrievalResult, error)
	RetrieveDocumentation(query string, maxResults int) ([]RetrievalResult, error)
	RetrieveConfigFiles(configType string, maxResults int) ([]RetrievalResult, error)

	// Context-aware retrieval
	RetrieveWithContext(query string, context []RetrievalResult, maxResults int) ([]RetrievalResult, error)
	RetrieveRelatedChunks(chunkID string, maxResults int) ([]RetrievalResult, error)

	// Filtering and ranking
	FilterResults(results []RetrievalResult, filters map[string]string) []RetrievalResult
	RerankResults(results []RetrievalResult, query string) ([]RetrievalResult, error)

	// Utilities
	GetRetrievalStats() *RetrievalStats
	ClearCache() error
}

// RetrievalStats represents statistics about retrieval operations
type RetrievalStats struct {
	TotalQueries       int                `json:"totalQueries"`       // Total number of queries
	CacheHits          int                `json:"cacheHits"`          // Number of cache hits
	CacheMisses        int                `json:"cacheMisses"`        // Number of cache misses
	AverageQueryTime   time.Duration      `json:"averageQueryTime"`   // Average query execution time
	AverageResultCount float64            `json:"averageResultCount"` // Average number of results returned
	MostCommonQueries  []QueryStats       `json:"mostCommonQueries"`  // Most common query patterns
	PerformanceMetrics map[string]float64 `json:"performanceMetrics"` // Various performance metrics
}

// QueryStats represents statistics for a specific query pattern
type QueryStats struct {
	QueryPattern string        `json:"queryPattern"` // Pattern or type of query
	Count        int           `json:"count"`        // Number of times queried
	AverageTime  time.Duration `json:"averageTime"`  // Average execution time
	AverageScore float32       `json:"averageScore"` // Average relevance score
}

// WikiPageContext represents context for wiki page generation
type WikiPageContext struct {
	PageTitle       string   `json:"pageTitle"`       // Title of the wiki page
	PageDescription string   `json:"pageDescription"` // Description of the page
	RequiredTopics  []string `json:"requiredTopics"`  // Topics that must be covered
	PreferredLang   []string `json:"preferredLang"`   // Preferred programming languages
	MaxSourceFiles  int      `json:"maxSourceFiles"`  // Maximum source files to include
	MinSourceFiles  int      `json:"minSourceFiles"`  // Minimum source files to include
	IncludeTests    bool     `json:"includeTests"`    // Whether to include test files
	IncludeConfigs  bool     `json:"includeConfigs"`  // Whether to include config files
	FocusAreas      []string `json:"focusAreas"`      // Specific areas to focus on
}

// SourceFileReference represents a reference to a source file for wiki generation
type SourceFileReference struct {
	FilePath          string            `json:"filePath"`          // Path to the source file
	Language          string            `json:"language"`          // Programming language
	Category          string            `json:"category"`          // File category
	Relevance         float32           `json:"relevance"`         // Relevance score to the page
	ImportantSections []string          `json:"importantSections"` // Important sections/functions
	LineReferences    []LineRef         `json:"lineReferences"`    // Specific line references
	Metadata          map[string]string `json:"metadata"`          // Additional metadata
}

// LineRef represents a reference to specific lines in a file
type LineRef struct {
	StartLine   int    `json:"startLine"`   // Starting line number
	EndLine     int    `json:"endLine"`     // Ending line number
	Description string `json:"description"` // Description of what these lines do
	Importance  int    `json:"importance"`  // Importance level (1-5)
}

// WikiContentRetriever specializes in retrieving content for wiki generation
type WikiContentRetriever interface {
	// Wiki-specific retrieval methods
	RetrieveForWikiPage(context *WikiPageContext) ([]SourceFileReference, error)
	RetrieveArchitectureInfo(projectPath string) ([]RetrievalResult, error)
	RetrieveAPIDocumentation(projectPath string) ([]RetrievalResult, error)
	RetrieveSetupInstructions(projectPath string) ([]RetrievalResult, error)

	// Content organization
	OrganizeContentByTopic(results []RetrievalResult) map[string][]RetrievalResult
	ExtractCodeExamples(results []RetrievalResult, language string) []CodeExample
	GenerateSourceFileList(results []RetrievalResult, minRelevance float32) []SourceFileReference

	// Context building
	BuildPageContext(pageTitle string, results []RetrievalResult) *WikiPageContext
	EnrichWithRelatedContent(context *WikiPageContext) error
}

// CodeExample represents a code example extracted from source files
type CodeExample struct {
	Language    string   `json:"language"`    // Programming language
	Code        string   `json:"code"`        // Code snippet
	Description string   `json:"description"` // Description of the code
	FilePath    string   `json:"filePath"`    // Source file path
	LineStart   int      `json:"lineStart"`   // Starting line number
	LineEnd     int      `json:"lineEnd"`     // Ending line number
	Context     string   `json:"context"`     // Surrounding context
	Tags        []string `json:"tags"`        // Tags/keywords
	Complexity  int      `json:"complexity"`  // Complexity score (1-5)
}

// RetrievalCache interface for caching retrieval results
type RetrievalCache interface {
	Get(key string) ([]RetrievalResult, bool)
	Set(key string, results []RetrievalResult, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
	Stats() map[string]interface{}
}
