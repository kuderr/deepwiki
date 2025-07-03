package processor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/scanner"
)

func TestNewTextProcessor(t *testing.T) {
	// Test with nil options (should use defaults)
	tp := NewTextProcessor(nil)
	if tp == nil {
		t.Fatal("Expected non-nil processor")
	}
	if tp.options == nil {
		t.Fatal("Expected default options to be set")
	}

	// Test with custom options
	options := &ProcessingOptions{
		ChunkSize:    200,
		ChunkOverlap: 50,
	}
	tp = NewTextProcessor(options)
	if tp.options.ChunkSize != 200 {
		t.Errorf("Expected chunk size 200, got %d", tp.options.ChunkSize)
	}
}

func TestChunkText(t *testing.T) {
	options := DefaultProcessingOptions()
	options.MinChunkWords = 5 // Lower threshold for testing
	tp := NewTextProcessor(options)

	// Test empty content
	fileInfo := scanner.FileInfo{
		Path:     "test.go",
		Language: "Go",
		Category: "code",
	}

	chunks, err := tp.ChunkText("", fileInfo)
	if err != nil {
		t.Errorf("Expected no error for empty content, got %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty content, got %d", len(chunks))
	}

	// Test simple text
	content := "This is a simple test content that should be chunked properly. " +
		"It contains multiple sentences and should be split according to the chunk size. " +
		"The chunking algorithm should handle this gracefully and create appropriate chunks."

	chunks, err = tp.ChunkText(content, fileInfo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}

	// Verify chunk properties
	for i, chunk := range chunks {
		if chunk.ID == "" {
			t.Errorf("Chunk %d has empty ID", i)
		}
		if chunk.Text == "" {
			t.Errorf("Chunk %d has empty text", i)
		}
		if chunk.WordCount <= 0 {
			t.Errorf("Chunk %d has invalid word count: %d", i, chunk.WordCount)
		}
	}
}

func TestChunkTextWithGoCode(t *testing.T) {
	options := DefaultProcessingOptions()
	options.MinChunkWords = 5 // Lower threshold for testing
	tp := NewTextProcessor(options)

	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func add(a, b int) int {
	return a + b
}

type User struct {
	Name string
	Age  int
}

func (u User) String() string {
	return fmt.Sprintf("User{Name: %s, Age: %d}", u.Name, u.Age)
}`

	fileInfo := scanner.FileInfo{
		Path:     "main.go",
		Language: "Go",
		Category: "code",
	}

	chunks, err := tp.ChunkText(goCode, fileInfo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk for Go code")
	}

	// Check if semantic chunking worked (should have metadata indicating semantic=true)
	foundSemantic := false
	for _, chunk := range chunks {
		if chunk.Metadata["semantic"] == "true" {
			foundSemantic = true
			break
		}
	}
	if !foundSemantic {
		t.Error("Expected at least one semantic chunk for Go code")
	}
}

func TestProcessFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is a test file content for processing. It should be chunked properly. " +
		"Adding more content to ensure we have enough words for chunking to work correctly. " +
		"This should now meet the minimum word count requirements for creating chunks. " +
		"We need to have sufficient content so that the chunking algorithm can create meaningful chunks."

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	options := DefaultProcessingOptions()
	options.MinChunkWords = 5 // Lower threshold for testing
	tp := NewTextProcessor(options)

	fileInfo := scanner.FileInfo{
		Path:         "test.txt",
		AbsolutePath: testFile,
		Name:         "test.txt",
		Extension:    ".txt",
		Size:         int64(len(testContent)),
		ModTime:      time.Now(),
		IsText:       true,
		Language:     "Text",
		Category:     "docs",
	}

	doc, err := tp.ProcessFile(fileInfo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if doc == nil {
		t.Fatal("Expected non-nil document")
	}

	// Verify document properties
	if doc.ID == "" {
		t.Error("Document ID should not be empty")
	}
	if doc.FilePath != fileInfo.Path {
		t.Errorf("Expected file path %s, got %s", fileInfo.Path, doc.FilePath)
	}
	if doc.Content != testContent {
		t.Error("Document content doesn't match original")
	}
	if len(doc.Chunks) == 0 {
		t.Error("Document should have at least one chunk")
	}
}

func TestProcessFiles(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()

	files := []struct {
		name    string
		content string
		lang    string
		cat     string
	}{
		{
			"test1.go",
			"package main\nfunc main() { fmt.Println(\"Hello\") }\nfunc add(a, b int) int { return a + b }",
			"Go",
			"code",
		},
		{"test2.py", "def hello():\n    print('Hello World')\n\ndef add(a, b):\n    return a + b", "Python", "code"},
		{
			"readme.md",
			"# Test Project\nThis is a comprehensive test project for testing chunking functionality.",
			"Markdown",
			"docs",
		},
	}

	fileInfos := make([]scanner.FileInfo, len(files))
	for i, file := range files {
		testFile := filepath.Join(tempDir, file.name)
		err := os.WriteFile(testFile, []byte(file.content), 0o644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file.name, err)
		}

		fileInfos[i] = scanner.FileInfo{
			Path:         file.name,
			AbsolutePath: testFile,
			Name:         file.name,
			Extension:    filepath.Ext(file.name),
			Size:         int64(len(file.content)),
			ModTime:      time.Now(),
			IsText:       true,
			Language:     file.lang,
			Category:     file.cat,
		}
	}

	options := DefaultProcessingOptions()
	options.MinChunkWords = 5 // Lower threshold for testing
	tp := NewTextProcessor(options)

	result, err := tp.ProcessFiles(fileInfos)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Documents) != len(files) {
		t.Errorf("Expected %d documents, got %d", len(files), len(result.Documents))
	}

	if result.TotalFiles != len(files) {
		t.Errorf("Expected total files %d, got %d", len(files), result.TotalFiles)
	}

	if result.TotalChunks == 0 {
		t.Error("Expected at least some chunks")
	}
}

func TestPreprocessContent(t *testing.T) {
	tp := NewTextProcessor(&ProcessingOptions{
		NormalizeWhitespace: true,
		RemoveComments:      true,
	})

	// Test whitespace normalization
	content := "This   has    multiple   spaces\n\n\n\nand   extra   newlines"
	processed := tp.preprocessContent(content, "Text")

	if processed == content {
		t.Error("Content should have been preprocessed")
	}

	// Test comment removal for Go
	goContent := `package main
// This is a comment
func main() {
	/* This is a block comment */
	fmt.Println("Hello")
}`

	processed = tp.preprocessContent(goContent, "Go")
	if processed == goContent {
		t.Error("Go comments should have been removed")
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"hello   world", 2},
		{"hello\nworld\ttest", 3},
		{"   hello   world   ", 2},
	}

	for _, test := range tests {
		result := countWords(test.text)
		if result != test.expected {
			t.Errorf("countWords(%q) = %d, expected %d", test.text, result, test.expected)
		}
	}
}

func TestGetLanguageProcessor(t *testing.T) {
	// Test known language
	goProcessor := GetLanguageProcessor("Go")
	if goProcessor.Language != "Go" {
		t.Errorf("Expected language Go, got %s", goProcessor.Language)
	}
	if len(goProcessor.CommentPrefixes) == 0 {
		t.Error("Go processor should have comment prefixes")
	}
	if len(goProcessor.FunctionKeywords) == 0 {
		t.Error("Go processor should have function keywords")
	}

	// Test unknown language
	unknownProcessor := GetLanguageProcessor("Unknown")
	if unknownProcessor.Language != "Unknown" {
		t.Errorf("Expected language Unknown, got %s", unknownProcessor.Language)
	}
}

func TestDocumentHelpers(t *testing.T) {
	tp := NewTextProcessor(DefaultProcessingOptions())

	// Create test documents
	docs := []Document{
		{
			ID:       "doc1",
			FilePath: "test1.go",
			Language: "Go",
			Category: "code",
		},
		{
			ID:       "doc2",
			FilePath: "test2.py",
			Language: "Python",
			Category: "code",
		},
		{
			ID:       "doc3",
			FilePath: "readme.md",
			Language: "Markdown",
			Category: "docs",
		},
	}

	// Test GetDocumentByPath
	doc := tp.GetDocumentByPath(docs, "test1.go")
	if doc == nil {
		t.Error("Should find document by path")
	}
	if doc.ID != "doc1" {
		t.Errorf("Expected doc1, got %s", doc.ID)
	}

	// Test GetDocumentsByCategory
	codeDocs := tp.GetDocumentsByCategory(docs, "code")
	if len(codeDocs) != 2 {
		t.Errorf("Expected 2 code documents, got %d", len(codeDocs))
	}

	// Test GetDocumentsByLanguage
	goDocs := tp.GetDocumentsByLanguage(docs, "Go")
	if len(goDocs) != 1 {
		t.Errorf("Expected 1 Go document, got %d", len(goDocs))
	}

	// Test GetHighImportanceDocuments
	docs[0].Importance = 5
	docs[1].Importance = 3
	docs[2].Importance = 1

	highImportanceDocs := tp.GetHighImportanceDocuments(docs, 4)
	if len(highImportanceDocs) != 1 {
		t.Errorf("Expected 1 high importance document, got %d", len(highImportanceDocs))
	}
}

func TestProcessingOptions(t *testing.T) {
	options := DefaultProcessingOptions()

	if options.ChunkSize != 350 {
		t.Errorf("Expected default chunk size 350, got %d", options.ChunkSize)
	}
	if options.ChunkOverlap != 100 {
		t.Errorf("Expected default chunk overlap 100, got %d", options.ChunkOverlap)
	}
	if !options.NormalizeWhitespace {
		t.Error("Expected NormalizeWhitespace to be true by default")
	}
	if !options.CountTokens {
		t.Error("Expected CountTokens to be true by default")
	}
}

func TestTextChunkProperties(t *testing.T) {
	tp := NewTextProcessor(DefaultProcessingOptions())

	content := "This is a test content for verifying chunk properties. " +
		"It should generate chunks with proper metadata and token counts."

	fileInfo := scanner.FileInfo{
		Path:     "test.txt",
		Language: "Text",
		Category: "docs",
	}

	chunks, err := tp.ChunkText(content, fileInfo)
	if err != nil {
		t.Fatalf("Failed to chunk text: %v", err)
	}

	for i, chunk := range chunks {
		// Test ID generation
		if chunk.ID == "" {
			t.Errorf("Chunk %d has empty ID", i)
		}

		// Test word count
		actualWords := countWords(chunk.Text)
		if chunk.WordCount != actualWords {
			t.Errorf("Chunk %d word count mismatch: expected %d, got %d", i, actualWords, chunk.WordCount)
		}

		// Test token count (should be estimated if CountTokens is enabled)
		if tp.options.CountTokens && chunk.TokenCount <= 0 {
			t.Errorf("Chunk %d should have token count > 0", i)
		}

		// Test metadata
		if chunk.Metadata == nil {
			t.Errorf("Chunk %d should have metadata", i)
		}
	}
}

// Benchmark tests
func BenchmarkChunkText(b *testing.B) {
	tp := NewTextProcessor(DefaultProcessingOptions())

	// Create a reasonably large text for benchmarking
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is sentence number " + string(rune(i)) + " in the benchmark text. "
	}

	fileInfo := scanner.FileInfo{
		Path:     "benchmark.txt",
		Language: "Text",
		Category: "docs",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tp.ChunkText(content, fileInfo)
		if err != nil {
			b.Fatalf("Chunking failed: %v", err)
		}
	}
}

func BenchmarkProcessFiles(b *testing.B) {
	// Create temporary test files
	tempDir := b.TempDir()

	fileInfos := make([]scanner.FileInfo, 10)
	for i := 0; i < 10; i++ {
		content := "package main\nfunc test" + string(rune(i)) + "() { return }"
		testFile := filepath.Join(tempDir, "test"+string(rune(i))+".go")
		os.WriteFile(testFile, []byte(content), 0o644)

		fileInfos[i] = scanner.FileInfo{
			Path:         "test" + string(rune(i)) + ".go",
			AbsolutePath: testFile,
			Name:         "test" + string(rune(i)) + ".go",
			Extension:    ".go",
			Size:         int64(len(content)),
			ModTime:      time.Now(),
			IsText:       true,
			Language:     "Go",
			Category:     "code",
		}
	}

	tp := NewTextProcessor(DefaultProcessingOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tp.ProcessFiles(fileInfos)
		if err != nil {
			b.Fatalf("Processing failed: %v", err)
		}
	}
}
