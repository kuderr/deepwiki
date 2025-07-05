package processor

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kuderr/deepwiki/pkg/scanner"
)

// TextProcessor handles text processing and chunking
type TextProcessor struct {
	options *ProcessingOptions
}

// NewTextProcessor creates a new text processor with options
func NewTextProcessor(options *ProcessingOptions) *TextProcessor {
	if options == nil {
		options = DefaultProcessingOptions()
	}
	return &TextProcessor{
		options: options,
	}
}

// ProcessFiles processes multiple files and returns documents with chunks
func (tp *TextProcessor) ProcessFiles(files []scanner.FileInfo) (*ProcessingResult, error) {
	startTime := time.Now()

	result := &ProcessingResult{
		Documents:  make([]Document, 0, len(files)),
		TotalFiles: len(files),
		Errors:     make([]string, 0),
	}

	if tp.options.Concurrent {
		return tp.processFilesConcurrent(files, result, startTime)
	}

	return tp.processFilesSequential(files, result, startTime)
}

// processFilesSequential processes files one by one
func (tp *TextProcessor) processFilesSequential(
	files []scanner.FileInfo,
	result *ProcessingResult,
	startTime time.Time,
) (*ProcessingResult, error) {
	for _, file := range files {
		if file.IsDir || file.IsBinary || !file.IsText {
			continue
		}

		doc, err := tp.ProcessFile(file)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error processing %s: %v", file.Path, err))
			continue
		}

		if doc != nil {
			result.Documents = append(result.Documents, *doc)
			result.TotalChunks += len(doc.Chunks)
			for _, chunk := range doc.Chunks {
				result.TotalTokens += chunk.TokenCount
			}
		}
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// processFilesConcurrent processes files concurrently
func (tp *TextProcessor) processFilesConcurrent(
	files []scanner.FileInfo,
	result *ProcessingResult,
	startTime time.Time,
) (*ProcessingResult, error) {
	// Filter files first
	validFiles := make([]scanner.FileInfo, 0, len(files))
	for _, file := range files {
		if !file.IsDir && !file.IsBinary && file.IsText {
			validFiles = append(validFiles, file)
		}
	}

	maxWorkers := tp.options.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = 4
	}

	jobs := make(chan scanner.FileInfo, len(validFiles))
	results := make(chan *Document, len(validFiles))
	errors := make(chan error, len(validFiles))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				doc, err := tp.ProcessFile(file)
				if err != nil {
					errors <- fmt.Errorf("error processing %s: %v", file.Path, err)
					continue
				}
				if doc != nil {
					results <- doc
				}
			}
		}()
	}

	// Send jobs
	for _, file := range validFiles {
		jobs <- file
	}
	close(jobs)

	// Wait for completion and close channels
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect results
	documents := make([]Document, 0, len(validFiles))
	errorMsgs := make([]string, 0)
	totalChunks := 0
	totalTokens := 0

	for doc := range results {
		documents = append(documents, *doc)
		totalChunks += len(doc.Chunks)
		for _, chunk := range doc.Chunks {
			totalTokens += chunk.TokenCount
		}
	}

	for err := range errors {
		errorMsgs = append(errorMsgs, err.Error())
	}

	result.Documents = documents
	result.TotalChunks = totalChunks
	result.TotalTokens = totalTokens
	result.Errors = errorMsgs
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// ProcessFile processes a single file and returns a document with chunks
func (tp *TextProcessor) ProcessFile(fileInfo scanner.FileInfo) (*Document, error) {
	content, err := tp.readFileContent(fileInfo.AbsolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", fileInfo.Path, err)
	}

	if len(content) == 0 {
		return nil, nil // Skip empty files
	}

	// Create document
	doc := &Document{
		ID:          tp.generateDocumentID(fileInfo.Path),
		FilePath:    fileInfo.Path,
		Language:    fileInfo.Language,
		Category:    fileInfo.Category,
		Content:     string(content),
		ProcessedAt: time.Now(),
		Size:        fileInfo.Size,
		LineCount:   fileInfo.LineCount,
		Importance:  fileInfo.Importance,
		Metadata:    make(map[string]string),
	}

	// Add metadata
	doc.Metadata["absolutePath"] = fileInfo.AbsolutePath
	doc.Metadata["extension"] = fileInfo.Extension
	doc.Metadata["modTime"] = fileInfo.ModTime.Format(time.RFC3339)

	// Preprocess content
	processedContent := tp.preprocessContent(string(content), fileInfo.Language)

	// Create chunks
	chunks, err := tp.ChunkText(processedContent, fileInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk text: %v", err)
	}

	doc.Chunks = chunks
	return doc, nil
}

// ChunkText splits text into chunks based on the configured strategy
func (tp *TextProcessor) ChunkText(content string, fileInfo scanner.FileInfo) ([]TextChunk, error) {
	if len(content) == 0 {
		return nil, nil
	}

	// Get language-specific processor
	langProcessor := GetLanguageProcessor(fileInfo.Language)

	// Try semantic chunking first for code files
	if fileInfo.Category == string(scanner.CategoryCode) {
		if chunks := tp.chunkBySemanticBoundaries(content, langProcessor, fileInfo); len(chunks) > 0 {
			return chunks, nil
		}
	}

	// Fall back to word-based chunking
	return tp.chunkByWords(content, fileInfo), nil
}

// chunkBySemanticBoundaries attempts to chunk code by semantic boundaries
func (tp *TextProcessor) chunkBySemanticBoundaries(
	content string,
	langProcessor *LanguageSpecificProcessor,
	fileInfo scanner.FileInfo,
) []TextChunk {
	lines := strings.Split(content, "\n")
	chunks := make([]TextChunk, 0)

	currentChunk := make([]string, 0)
	currentPos := 0
	chunkID := 0

	for i, line := range lines {
		// Check if line starts a new semantic boundary
		isNewBoundary := false
		for _, boundary := range langProcessor.ChunkBoundaries {
			if strings.HasPrefix(strings.TrimSpace(line), boundary) {
				isNewBoundary = true
				break
			}
		}

		// If we hit a boundary and have content, finalize current chunk
		if isNewBoundary && len(currentChunk) > 0 {
			chunkText := strings.Join(currentChunk, "\n")
			wordCount := countWords(chunkText)

			if wordCount >= tp.options.MinChunkWords {
				chunk := TextChunk{
					ID:        fmt.Sprintf("%s_chunk_%d", tp.generateDocumentID(fileInfo.Path), chunkID),
					Text:      chunkText,
					WordCount: wordCount,
					StartPos:  currentPos,
					EndPos:    currentPos + len(chunkText),
					Metadata: map[string]string{
						"semantic":  "true",
						"startLine": fmt.Sprintf("%d", currentPos),
						"endLine":   fmt.Sprintf("%d", i),
					},
				}

				if tp.options.CountTokens {
					chunk.TokenCount = tp.estimateTokenCount(chunkText)
				}

				chunks = append(chunks, chunk)
				chunkID++
			}

			currentPos += len(chunkText) + 1
			currentChunk = make([]string, 0)
		}

		currentChunk = append(currentChunk, line)

		// Check if chunk is getting too large
		if len(currentChunk) > 0 {
			chunkText := strings.Join(currentChunk, "\n")
			if countWords(chunkText) > tp.options.MaxChunkWords {
				chunk := TextChunk{
					ID:        fmt.Sprintf("%s_chunk_%d", tp.generateDocumentID(fileInfo.Path), chunkID),
					Text:      chunkText,
					WordCount: countWords(chunkText),
					StartPos:  currentPos,
					EndPos:    currentPos + len(chunkText),
					Metadata: map[string]string{
						"semantic":  "true",
						"truncated": "true",
					},
				}

				if tp.options.CountTokens {
					chunk.TokenCount = tp.estimateTokenCount(chunkText)
				}

				chunks = append(chunks, chunk)
				chunkID++
				currentPos += len(chunkText) + 1
				currentChunk = make([]string, 0)
			}
		}
	}

	// Handle remaining content
	if len(currentChunk) > 0 {
		chunkText := strings.Join(currentChunk, "\n")
		wordCount := countWords(chunkText)

		if wordCount >= tp.options.MinChunkWords {
			chunk := TextChunk{
				ID:        fmt.Sprintf("%s_chunk_%d", tp.generateDocumentID(fileInfo.Path), chunkID),
				Text:      chunkText,
				WordCount: wordCount,
				StartPos:  currentPos,
				EndPos:    currentPos + len(chunkText),
				Metadata: map[string]string{
					"semantic": "true",
					"final":    "true",
				},
			}

			if tp.options.CountTokens {
				chunk.TokenCount = tp.estimateTokenCount(chunkText)
			}

			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// chunkByWords splits content into word-based chunks with overlap
func (tp *TextProcessor) chunkByWords(content string, fileInfo scanner.FileInfo) []TextChunk {
	words := strings.Fields(content)
	if len(words) == 0 {
		return nil
	}

	chunks := make([]TextChunk, 0)
	chunkSize := tp.options.ChunkSize
	overlap := tp.options.ChunkOverlap
	chunkID := 0

	for i := 0; i < len(words); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunkWords := words[i:end]
		chunkText := strings.Join(chunkWords, " ")

		// Skip if chunk is too small
		if len(chunkWords) < tp.options.MinChunkWords {
			break
		}

		// Calculate positions (approximate)
		startPos := i
		endPos := end - 1

		chunk := TextChunk{
			ID:        fmt.Sprintf("%s_chunk_%d", tp.generateDocumentID(fileInfo.Path), chunkID),
			Text:      chunkText,
			WordCount: len(chunkWords),
			StartPos:  startPos,
			EndPos:    endPos,
			Metadata: map[string]string{
				"chunkType": "word_based",
				"wordStart": fmt.Sprintf("%d", i),
				"wordEnd":   fmt.Sprintf("%d", end-1),
			},
		}

		if tp.options.CountTokens {
			chunk.TokenCount = tp.estimateTokenCount(chunkText)
		}

		chunks = append(chunks, chunk)
		chunkID++

		// Check max chunks limit
		if tp.options.MaxChunks > 0 && len(chunks) >= tp.options.MaxChunks {
			break
		}

		// If we've processed all words, break
		if end >= len(words) {
			break
		}
	}

	return chunks
}

// preprocessContent applies preprocessing to content based on options
func (tp *TextProcessor) preprocessContent(content, language string) string {
	if tp.options.NormalizeWhitespace {
		content = tp.normalizeWhitespace(content)
	}

	if tp.options.RemoveComments {
		content = tp.removeComments(content, language)
	}

	return content
}

// normalizeWhitespace normalizes whitespace in content
func (tp *TextProcessor) normalizeWhitespace(content string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	content = re.ReplaceAllString(content, " ")

	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Remove excessive blank lines (more than 2 consecutive)
	re = regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}

// removeComments removes comments from code content
func (tp *TextProcessor) removeComments(content, language string) string {
	langProcessor := GetLanguageProcessor(language)

	// Remove single-line comments
	for _, prefix := range langProcessor.CommentPrefixes {
		re := regexp.MustCompile(regexp.QuoteMeta(prefix) + `.*$`)
		content = re.ReplaceAllString(content, "")
	}

	// Remove multi-line comments
	for _, block := range langProcessor.CommentBlocks {
		if len(block) == 2 {
			start := regexp.QuoteMeta(block[0])
			end := regexp.QuoteMeta(block[1])
			re := regexp.MustCompile(start + `.*?` + end)
			content = re.ReplaceAllString(content, "")
		}
	}

	return content
}

// detectContentType determines the content type based on file extension and path
func (tp *TextProcessor) detectContentType(filePath string) ContentType {
	ext := filepath.Ext(filePath)
	fileName := strings.ToLower(filepath.Base(filePath))

	// Check for test files first (by naming pattern)
	if strings.Contains(fileName, "test") || strings.Contains(fileName, "spec") {
		return ContentTypeTest
	}

	// Check by extension
	switch ext {
	case ".md", ".txt", ".rst", ".adoc", ".org":
		return ContentTypeDocumentation
	case ".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf", ".xml", ".env", ".properties":
		return ContentTypeConfiguration
	case ".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".hpp", ".rs", ".jsx", ".tsx",
		".php", ".swift", ".cs", ".rb", ".kt", ".scala", ".clj", ".hs", ".ml", ".fs", ".elm",
		".dart", ".jl", ".html", ".css", ".scss", ".sass", ".less", ".vue", ".svelte", ".sh",
		".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd":
		return ContentTypeCode
	case ".csv", ".tsv", ".sql", ".db", ".sqlite", ".sqlite3":
		return ContentTypeData
	default:
		// Check for special files without extensions
		switch fileName {
		case "dockerfile", "makefile", "rakefile", "gemfile", "guardfile":
			return ContentTypeConfiguration
		default:
			return ContentTypeUnknown
		}
	}
}

// readFileContent reads file content and handles encoding
func (tp *TextProcessor) readFileContent(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check file size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Determine content type and get appropriate size limit
	contentType := tp.detectContentType(filePath)
	sizeLimit := tp.options.MaxFileSizeLimits[contentType]

	// Fall back to default if content type not configured
	if sizeLimit == 0 {
		sizeLimit = int64(tp.options.MaxChunkWords * 10)
	}

	if info.Size() > sizeLimit {
		return nil, fmt.Errorf("file too large: %d bytes (limit: %d bytes for %s files)",
			info.Size(), sizeLimit, contentType)
	}

	// Read content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Validate UTF-8
	if !utf8.Valid(content) {
		// Try to clean invalid UTF-8
		content = []byte(strings.ToValidUTF8(string(content), ""))
	}

	return content, nil
}

// generateDocumentID generates a unique ID for a document
func (tp *TextProcessor) generateDocumentID(filePath string) string {
	hash := md5.Sum([]byte(filePath))
	return fmt.Sprintf("doc_%x", hash[:8])
}

// estimateTokenCount provides a rough token count estimation
func (tp *TextProcessor) estimateTokenCount(text string) int {
	// TODO: Integrate with tiktoken-go for accurate token counting
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// countWords counts words in text
func countWords(text string) int {
	if len(text) == 0 {
		return 0
	}

	words := 0
	inWord := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}

	return words
}

// GetDocumentByPath finds a document by file path
func (tp *TextProcessor) GetDocumentByPath(documents []Document, filePath string) *Document {
	for i := range documents {
		if documents[i].FilePath == filePath {
			return &documents[i]
		}
	}
	return nil
}

// GetDocumentsByCategory returns documents filtered by category
func (tp *TextProcessor) GetDocumentsByCategory(documents []Document, category string) []Document {
	result := make([]Document, 0)
	for _, doc := range documents {
		if doc.Category == category {
			result = append(result, doc)
		}
	}
	return result
}

// GetDocumentsByLanguage returns documents filtered by programming language
func (tp *TextProcessor) GetDocumentsByLanguage(documents []Document, language string) []Document {
	result := make([]Document, 0)
	for _, doc := range documents {
		if doc.Language == language {
			result = append(result, doc)
		}
	}
	return result
}

// GetHighImportanceDocuments returns documents with high importance scores
func (tp *TextProcessor) GetHighImportanceDocuments(documents []Document, minImportance int) []Document {
	result := make([]Document, 0)
	for _, doc := range documents {
		if doc.Importance >= minImportance {
			result = append(result, doc)
		}
	}
	return result
}
