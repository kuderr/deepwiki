package generator

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/llm"
	"github.com/kuderr/deepwiki/pkg/rag"
	"github.com/kuderr/deepwiki/pkg/scanner"
)

// MockLLMProvider implements the llm.Provider interface for testing
type MockLLMProvider struct{}

func (m *MockLLMProvider) ChatCompletion(
	ctx context.Context,
	messages []llm.Message,
	opts ...llm.ChatCompletionOptions,
) (*llm.ChatCompletionResponse, error) {
	return &llm.ChatCompletionResponse{
		Choices: []llm.Choice{
			{Message: llm.Message{Content: "test response"}},
		},
	}, nil
}

func (m *MockLLMProvider) ChatCompletionStream(
	ctx context.Context,
	messages []llm.Message,
	handler llm.StreamHandler,
	opts ...llm.ChatCompletionOptions,
) error {
	return nil
}

func (m *MockLLMProvider) CountTokens(text string) (int, error) {
	return len(strings.Fields(text)), nil
}

func (m *MockLLMProvider) EstimateCost(promptTokens, completionTokens int) float64 {
	return 0.0
}

func (m *MockLLMProvider) GetUsageStats() llm.TokenCount {
	return llm.TokenCount{}
}

func (m *MockLLMProvider) ResetUsageStats() {}

func (m *MockLLMProvider) GetProviderType() llm.ProviderType {
	return llm.ProviderOpenAI
}

func (m *MockLLMProvider) GetModel() string {
	return "gpt-4o"
}

// MockRAGRetriever implements the rag.DocumentRetriever interface for testing
type MockRAGRetriever struct{}

func (m *MockRAGRetriever) RetrieveRelevantDocuments(ctx *rag.RetrievalContext) ([]rag.RetrievalResult, error) {
	return []rag.RetrievalResult{
		{FilePath: "test.go", Content: "package main"},
	}, nil
}

func (m *MockRAGRetriever) RetrieveByQuery(query string, maxResults int) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveByTags(tags []string, maxResults int) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveCodeExamples(
	language, concept string,
	maxResults int,
) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveDocumentation(query string, maxResults int) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveConfigFiles(configType string, maxResults int) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveWithContext(
	query string,
	context []rag.RetrievalResult,
	maxResults int,
) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) RetrieveRelatedChunks(chunkID string, maxResults int) ([]rag.RetrievalResult, error) {
	return nil, nil
}

func (m *MockRAGRetriever) FilterResults(
	results []rag.RetrievalResult,
	filters map[string]string,
) []rag.RetrievalResult {
	return results
}

func (m *MockRAGRetriever) RerankResults(results []rag.RetrievalResult, query string) ([]rag.RetrievalResult, error) {
	return results, nil
}

func (m *MockRAGRetriever) GetRetrievalStats() *rag.RetrievalStats {
	return nil
}

func TestNewWikiGenerator(t *testing.T) {
	client := &MockLLMProvider{}
	retriever := &MockRAGRetriever{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	generator := NewWikiGenerator(client, retriever, logger)

	if generator == nil {
		t.Error("Expected generator to be created")
	}
	if generator.xmlParser == nil {
		t.Error("Expected XML parser to be initialized")
	}
}

func TestBuildFileTree(t *testing.T) {
	generator := &WikiGenerator{}

	files := []scanner.FileInfo{
		{Path: "main.go"},
		{Path: "src/app.go"},
		{Path: "src/config/config.go"},
		{Path: "README.md"},
	}

	tree := generator.buildFileTree(files, "/test")

	if tree == "" {
		t.Error("Expected file tree to be generated")
	}

	// Check that tree contains expected files
	expectedFiles := []string{"main.go", "app.go", "config.go", "README.md"}
	for _, file := range expectedFiles {
		if !strings.Contains(tree, file) {
			t.Errorf("Expected file tree to contain %s", file)
		}
	}
}

func TestFindReadmeContent(t *testing.T) {
	generator := &WikiGenerator{}

	// Create temporary directory and README file
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	readmeContent := "# Test Project\n\nThis is a test README.md file for testing."

	err := os.WriteFile(readmePath, []byte(readmeContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test README file: %v", err)
	}

	files := []scanner.FileInfo{
		{Path: filepath.Join(tempDir, "main.go")},
		{Path: readmePath},
		{Path: filepath.Join(tempDir, "src/app.go")},
	}

	content := generator.findReadmeContent(files)

	if content != readmeContent {
		t.Errorf("Expected README content to be %q, got %q", readmeContent, content)
	}
}

func TestFindReadmeContentNotFound(t *testing.T) {
	generator := &WikiGenerator{}

	files := []scanner.FileInfo{
		{Path: "main.go"},
		{Path: "src/app.go"},
	}

	content := generator.findReadmeContent(files)

	if !strings.Contains(content, "No README file found") {
		t.Error("Expected message about missing README file")
	}
}

func TestFormatRelevantFiles(t *testing.T) {
	generator := &WikiGenerator{}

	docs := []rag.RetrievalResult{
		{
			FilePath: "main.go",
			Content:  "package main\n\nfunc main() {}",
		},
		{
			FilePath: "config.go",
			Content:  "package config\n\ntype Config struct {}",
		},
	}

	formatted := generator.formatRelevantFiles(docs)

	if !strings.Contains(formatted, "main.go") {
		t.Error("Expected formatted output to contain main.go")
	}
	if !strings.Contains(formatted, "config.go") {
		t.Error("Expected formatted output to contain config.go")
	}
	if !strings.Contains(formatted, "package main") {
		t.Error("Expected formatted output to contain file content")
	}
}

func TestGenerationOptions(t *testing.T) {
	options := GenerationOptions{
		ProjectName:    "test-project",
		ProjectPath:    "/test/path",
		Language:       "en",
		OutputFormat:   "markdown",
		MaxConcurrency: 2,
	}

	if options.ProjectName != "test-project" {
		t.Error("Expected project name to be set")
	}
	if options.Language != "en" {
		t.Error("Expected language to be set")
	}
}

func TestNoOpProgressTracker(t *testing.T) {
	tracker := &NoOpProgressTracker{}

	// These should not panic
	tracker.StartTask("test", 10)
	tracker.UpdateProgress(5, "halfway")
	tracker.CompleteTask("done")
	tracker.SetError(nil)
}

func TestGenerationStats(t *testing.T) {
	stats1 := GenerationStats{
		FilesProcessed:  10,
		PagesGenerated:  5,
		TotalTokensUsed: 1000,
		StartTime:       time.Now(),
	}

	stats2 := GenerationStats{
		FilesProcessed:  5,
		PagesGenerated:  3,
		TotalTokensUsed: 500,
		EndTime:         time.Now(),
	}

	combined := stats1.Add(stats2)

	if combined.FilesProcessed != 15 {
		t.Errorf("Expected combined files processed to be 15, got %d", combined.FilesProcessed)
	}
	if combined.PagesGenerated != 8 {
		t.Errorf("Expected combined pages generated to be 8, got %d", combined.PagesGenerated)
	}
	if combined.TotalTokensUsed != 1500 {
		t.Errorf("Expected combined tokens used to be 1500, got %d", combined.TotalTokensUsed)
	}
}

func TestGenerationStatsDuration(t *testing.T) {
	start := time.Now()
	end := start.Add(5 * time.Second)

	stats := GenerationStats{
		StartTime: start,
		EndTime:   end,
	}

	duration := stats.Duration()
	if duration != 5*time.Second {
		t.Errorf("Expected duration to be 5 seconds, got %v", duration)
	}
}
