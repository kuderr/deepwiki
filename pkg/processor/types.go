package processor

import (
	"time"

	"github.com/kuderr/deepwiki/pkg/scanner"
)

// TextChunk represents a chunk of text with metadata
type TextChunk struct {
	ID         string            `json:"id"`         // Unique identifier for the chunk
	Text       string            `json:"text"`       // The actual text content
	WordCount  int               `json:"wordCount"`  // Number of words in the chunk
	TokenCount int               `json:"tokenCount"` // Estimated token count
	StartPos   int               `json:"startPos"`   // Starting position in original text
	EndPos     int               `json:"endPos"`     // Ending position in original text
	Metadata   map[string]string `json:"metadata"`   // Additional metadata
}

// Document represents a processed document with chunks and embeddings
type Document struct {
	ID          string            `json:"id"`          // Unique document identifier
	FilePath    string            `json:"filePath"`    // Original file path
	Language    string            `json:"language"`    // Programming language
	Category    string            `json:"category"`    // File category
	Content     string            `json:"content"`     // Full document content
	Chunks      []TextChunk       `json:"chunks"`      // Text chunks
	Metadata    map[string]string `json:"metadata"`    // Document metadata
	ProcessedAt time.Time         `json:"processedAt"` // When document was processed
	Size        int64             `json:"size"`        // File size in bytes
	LineCount   int               `json:"lineCount"`   // Number of lines
	Importance  int               `json:"importance"`  // Importance score (1-5)
}

// ProcessingOptions represents options for text processing
type ProcessingOptions struct {
	// Chunking options
	ChunkSize    int `json:"chunkSize"`    // Target chunk size in words (default: 350)
	ChunkOverlap int `json:"chunkOverlap"` // Overlap between chunks in words (default: 100)
	MaxChunks    int `json:"maxChunks"`    // Maximum chunks per document (0 = unlimited)

	// Content preprocessing
	RemoveComments      bool `json:"removeComments"`      // Remove code comments
	NormalizeWhitespace bool `json:"normalizeWhitespace"` // Normalize whitespace
	PreserveStructure   bool `json:"preserveStructure"`   // Preserve code structure

	// Filtering options
	MinChunkWords   int  `json:"minChunkWords"`   // Minimum words per chunk (default: 50)
	MaxChunkWords   int  `json:"maxChunkWords"`   // Maximum words per chunk (default: 500)
	SkipEmptyChunks bool `json:"skipEmptyChunks"` // Skip chunks with no meaningful content

	// Token counting
	CountTokens bool   `json:"countTokens"` // Count tokens for each chunk
	TokenModel  string `json:"tokenModel"`  // Model to use for token counting (default: "cl100k_base")

	// Performance options
	Concurrent bool `json:"concurrent"` // Process documents concurrently
	MaxWorkers int  `json:"maxWorkers"` // Maximum worker goroutines
}

// DefaultProcessingOptions returns default processing options
func DefaultProcessingOptions() *ProcessingOptions {
	return &ProcessingOptions{
		ChunkSize:           350,
		ChunkOverlap:        100,
		MaxChunks:           0, // unlimited
		RemoveComments:      false,
		NormalizeWhitespace: true,
		PreserveStructure:   true,
		MinChunkWords:       50,
		MaxChunkWords:       500,
		SkipEmptyChunks:     true,
		CountTokens:         true,
		TokenModel:          "cl100k_base",
		Concurrent:          true,
		MaxWorkers:          4,
	}
}

// ProcessingResult represents the result of document processing
type ProcessingResult struct {
	Documents      []Document    `json:"documents"`      // Processed documents
	TotalFiles     int           `json:"totalFiles"`     // Total files processed
	TotalChunks    int           `json:"totalChunks"`    // Total chunks created
	TotalTokens    int           `json:"totalTokens"`    // Total tokens counted
	ProcessingTime time.Duration `json:"processingTime"` // Time taken to process
	Errors         []string      `json:"errors"`         // Any errors encountered
}

// ChunkingStrategy represents different strategies for text chunking
type ChunkingStrategy string

const (
	StrategyWordBased      ChunkingStrategy = "word_based"      // Chunk by word count
	StrategySentenceBased  ChunkingStrategy = "sentence_based"  // Chunk by sentences
	StrategyParagraphBased ChunkingStrategy = "paragraph_based" // Chunk by paragraphs
	StrategySemanticBased  ChunkingStrategy = "semantic_based"  // Chunk by semantic boundaries
)

// ContentType represents the type of content in a document
type ContentType string

const (
	ContentTypeCode          ContentType = "code"          // Source code
	ContentTypeDocumentation ContentType = "documentation" // Documentation files
	ContentTypeConfiguration ContentType = "configuration" // Config files
	ContentTypeData          ContentType = "data"          // Data files
	ContentTypeTest          ContentType = "test"          // Test files
	ContentTypeUnknown       ContentType = "unknown"       // Unknown content type
)

// LanguageSpecificProcessor represents language-specific processing rules
type LanguageSpecificProcessor struct {
	Language         string     `json:"language"`         // Programming language
	CommentPrefixes  []string   `json:"commentPrefixes"`  // Single-line comment prefixes
	CommentBlocks    [][]string `json:"commentBlocks"`    // Multi-line comment blocks [start, end]
	StringDelimiters []string   `json:"stringDelimiters"` // String delimiters
	ImportKeywords   []string   `json:"importKeywords"`   // Import/include keywords
	FunctionKeywords []string   `json:"functionKeywords"` // Function definition keywords
	ClassKeywords    []string   `json:"classKeywords"`    // Class definition keywords
	ChunkBoundaries  []string   `json:"chunkBoundaries"`  // Natural chunk boundaries
}

// GetLanguageProcessor returns language-specific processing rules
func GetLanguageProcessor(language string) *LanguageSpecificProcessor {
	processors := map[string]*LanguageSpecificProcessor{
		"Go": {
			Language:         "Go",
			CommentPrefixes:  []string{"//"},
			CommentBlocks:    [][]string{{"/*", "*/"}},
			StringDelimiters: []string{`"`, "`"},
			ImportKeywords:   []string{"import", "package"},
			FunctionKeywords: []string{"func"},
			ClassKeywords:    []string{"type", "struct", "interface"},
			ChunkBoundaries:  []string{"func ", "type ", "var ", "const "},
		},
		"Python": {
			Language:         "Python",
			CommentPrefixes:  []string{"#"},
			CommentBlocks:    [][]string{{`"""`, `"""`}, {`'''`, `'''`}},
			StringDelimiters: []string{`"`, `'`, `"""`, `'''`},
			ImportKeywords:   []string{"import", "from"},
			FunctionKeywords: []string{"def "},
			ClassKeywords:    []string{"class "},
			ChunkBoundaries:  []string{"def ", "class ", "if __name__"},
		},
		"JavaScript": {
			Language:         "JavaScript",
			CommentPrefixes:  []string{"//"},
			CommentBlocks:    [][]string{{"/*", "*/"}},
			StringDelimiters: []string{`"`, `'`, "`"},
			ImportKeywords:   []string{"import", "require", "export"},
			FunctionKeywords: []string{"function ", "async function", "const ", "let ", "var "},
			ClassKeywords:    []string{"class "},
			ChunkBoundaries:  []string{"function ", "class ", "export ", "module.exports"},
		},
		"TypeScript": {
			Language:         "TypeScript",
			CommentPrefixes:  []string{"//"},
			CommentBlocks:    [][]string{{"/*", "*/"}},
			StringDelimiters: []string{`"`, `'`, "`"},
			ImportKeywords:   []string{"import", "export"},
			FunctionKeywords: []string{"function ", "async function", "const ", "let ", "var "},
			ClassKeywords:    []string{"class ", "interface ", "type "},
			ChunkBoundaries:  []string{"function ", "class ", "interface ", "type ", "export "},
		},
		"Java": {
			Language:         "Java",
			CommentPrefixes:  []string{"//"},
			CommentBlocks:    [][]string{{"/*", "*/"}, {"/**", "*/"}},
			StringDelimiters: []string{`"`},
			ImportKeywords:   []string{"import", "package"},
			FunctionKeywords: []string{"public ", "private ", "protected "},
			ClassKeywords:    []string{"class ", "interface ", "enum "},
			ChunkBoundaries:  []string{"class ", "interface ", "enum ", "public ", "private "},
		},
	}

	if processor, exists := processors[language]; exists {
		return processor
	}

	// Return default processor for unknown languages
	return &LanguageSpecificProcessor{
		Language:         language,
		CommentPrefixes:  []string{"//", "#"},
		CommentBlocks:    [][]string{{"/*", "*/"}},
		StringDelimiters: []string{`"`, `'`},
		ImportKeywords:   []string{},
		FunctionKeywords: []string{},
		ClassKeywords:    []string{},
		ChunkBoundaries:  []string{},
	}
}

// FileProcessor interface for processing different file types
type FileProcessor interface {
	ProcessFile(fileInfo scanner.FileInfo, content []byte, options *ProcessingOptions) (*Document, error)
	CanProcess(fileInfo scanner.FileInfo) bool
	GetContentType(fileInfo scanner.FileInfo) ContentType
}
