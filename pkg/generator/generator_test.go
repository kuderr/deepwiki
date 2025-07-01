package generator

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/openai"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/rag"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/scanner"
)

// MockOpenAIClient implements the openai.Client interface for testing
type MockOpenAIClient struct{}

func (m *MockOpenAIClient) ChatCompletion(
	ctx context.Context,
	messages []openai.Message,
	opts ...openai.ChatCompletionOptions,
) (*openai.ChatCompletionResponse, error) {
	return &openai.ChatCompletionResponse{
		Choices: []openai.Choice{
			{Message: openai.Message{Content: "test response"}},
		},
	}, nil
}

func (m *MockOpenAIClient) ChatCompletionStream(
	ctx context.Context,
	messages []openai.Message,
	handler openai.StreamHandler,
	opts ...openai.ChatCompletionOptions,
) error {
	return nil
}

func (m *MockOpenAIClient) CreateEmbeddings(
	ctx context.Context,
	texts []string,
	opts ...openai.EmbeddingOptions,
) (*openai.EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockOpenAIClient) CountTokens(text string) (int, error) {
	return len(strings.Fields(text)), nil
}

func (m *MockOpenAIClient) EstimateCost(promptTokens, completionTokens int) float64 {
	return 0.0
}

func (m *MockOpenAIClient) GetUsageStats() openai.TokenCount {
	return openai.TokenCount{}
}

func (m *MockOpenAIClient) ResetUsageStats() {}

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

func (m *MockRAGRetriever) ClearCache() error {
	return nil
}

func TestNewWikiGenerator(t *testing.T) {
	client := &MockOpenAIClient{}
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

	files := []scanner.FileInfo{
		{Path: "main.go"},
		{Path: "README.md"},
		{Path: "src/app.go"},
	}

	content := generator.findReadmeContent(files)

	if !strings.Contains(content, "README.md") {
		t.Error("Expected README content to mention README.md file")
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
