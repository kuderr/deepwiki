package generator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuderr/deepwiki/internal/prompts"
	"github.com/kuderr/deepwiki/pkg/openai"
	"github.com/kuderr/deepwiki/pkg/rag"
	"github.com/kuderr/deepwiki/pkg/scanner"
)

// WikiGenerator generates wiki structures and content
type WikiGenerator struct {
	openaiClient openai.Client
	ragRetriever rag.DocumentRetriever
	xmlParser    *XMLParser
	logger       *slog.Logger
}

// NewWikiGenerator creates a new wiki generator
func NewWikiGenerator(client openai.Client, retriever rag.DocumentRetriever, logger *slog.Logger) *WikiGenerator {
	return &WikiGenerator{
		openaiClient: client,
		ragRetriever: retriever,
		xmlParser:    NewXMLParser(),
		logger:       logger.With("component", "generator"),
	}
}

// GenerateWikiStructure generates the overall wiki structure for a project
func (g *WikiGenerator) GenerateWikiStructure(
	ctx context.Context,
	fileTree string,
	readmeContent string,
	options GenerationOptions,
) (*WikiStructure, error) {
	g.logger.Info("Starting wiki structure generation",
		"project", options.ProjectName,
		"fileTree", fileTree,
		"language", options.Language)

	start := time.Now()

	// Prepare prompt data
	promptData := prompts.WikiStructureData{
		FileTree:    fileTree,
		ReadmeFile:  readmeContent,
		ProjectName: options.ProjectName,
		Language:    options.Language,
	}

	// Execute the prompt
	prompt, err := prompts.ExecuteWikiStructurePrompt(promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate structure prompt: %w", err)
	}

	g.logger.Debug("Generated structure prompt", "length", len(prompt))

	// Call OpenAI API
	messages := []openai.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := g.openaiClient.ChatCompletion(ctx, messages, openai.ChatCompletionOptions{
		MaxTokens:   4000,
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API for structure generation: %w", err)
	}

	// Parse the XML response
	structureResponse, err := g.xmlParser.ParseWikiStructure(response.Choices[0].Message.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wiki structure response: %w", err)
	}

	// Convert to WikiStructure
	structure := g.xmlParser.ConvertToWikiStructure(structureResponse, options)

	g.logger.Info("Wiki structure generated successfully",
		"pages", len(structure.Pages),
		"duration", time.Since(start))

	return structure, nil
}

// GeneratePageContent generates content for a specific wiki page
func (g *WikiGenerator) GeneratePageContent(
	ctx context.Context,
	fileTree string,
	page *WikiPage,
	structure *WikiStructure,
	options GenerationOptions,
) error {
	g.logger.Info("Generating content for page", "page", page.Title, "id", page.ID)

	start := time.Now()

	// Retrieve relevant documents for this page using a simple query based on page title
	retrievalContext := &rag.RetrievalContext{
		Query:      page.Title + " " + page.Description,
		QueryType:  rag.QueryTypeHybrid,
		MaxResults: 20,
		MinScore:   0.1,
	}

	relevantDocs, err := g.ragRetriever.RetrieveRelevantDocuments(retrievalContext)
	if err != nil {
		return fmt.Errorf("failed to retrieve relevant documents for page %s: %w", page.ID, err)
	}

	g.logger.Debug("Retrieved relevant documents", "page", page.ID, "docs", len(relevantDocs))

	// Format relevant files for the prompt
	relevantFiles := g.formatRelevantFiles(relevantDocs)

	// Prepare prompt data
	promptData := prompts.PageContentData{
		PageTitle:       page.Title,
		PageDescription: page.Description,
		RelevantFiles:   relevantFiles,
		ProjectName:     options.ProjectName,
		Language:        options.Language,
		FileTree:        fileTree,
	}

	// Execute the prompt
	prompt, err := prompts.ExecutePageContentPrompt(promptData)
	if err != nil {
		return fmt.Errorf("failed to generate content prompt for page %s: %w", page.ID, err)
	}

	// Call OpenAI API
	messages := []openai.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := g.openaiClient.ChatCompletion(ctx, messages, openai.ChatCompletionOptions{
		MaxTokens:   4000,
		Temperature: 0.1,
	})
	if err != nil {
		return fmt.Errorf("failed to call OpenAI API for page content generation: %w", err)
	}

	// Update the page with generated content
	page.Content = response.Choices[0].Message.Content
	page.WordCount = len(strings.Fields(response.Choices[0].Message.Content))
	page.SourceFiles = len(relevantDocs)
	page.CreatedAt = time.Now()

	// Extract file paths from relevant documents
	filePaths := make([]string, len(relevantDocs))
	for i, doc := range relevantDocs {
		filePaths[i] = doc.FilePath
	}
	page.FilePaths = filePaths

	g.logger.Info("Page content generated successfully",
		"page", page.ID,
		"words", page.WordCount,
		"sources", page.SourceFiles,
		"duration", time.Since(start))

	return nil
}

// GenerateWiki generates a complete wiki for the project
func (g *WikiGenerator) GenerateWiki(
	ctx context.Context,
	files []scanner.FileInfo,
	options GenerationOptions,
) (*GenerationResult, error) {
	result := &GenerationResult{
		GeneratedAt: time.Now(),
		Pages:       make(map[string]*WikiPage),
	}

	start := time.Now()
	defer func() {
		result.ProcessingTime = time.Since(start)
	}()

	// Initialize progress tracker
	if options.ProgressTracker == nil {
		options.ProgressTracker = &NoOpProgressTracker{}
	}

	fileTree := g.buildFileTree(files, options.ProjectPath)
	readmeContent := g.findReadmeContent(files)

	// Step 1: Generate wiki structure
	options.ProgressTracker.StartTask("Generating wiki structure", 1)
	structure, err := g.GenerateWikiStructure(ctx, fileTree, readmeContent, options)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("structure generation failed: %w", err))
		return result, err
	}
	result.Structure = structure
	options.ProgressTracker.CompleteTask("Wiki structure generated")

	// Step 2: Generate content for each page
	options.ProgressTracker.StartTask("Generating page content", len(structure.Pages))

	for i, page := range structure.Pages {
		pagePtr := &structure.Pages[i] // Get pointer to the actual page in the slice

		options.ProgressTracker.UpdateProgress(i, fmt.Sprintf("Generating: %s", page.Title))

		if err := g.GeneratePageContent(ctx, fileTree, pagePtr, structure, options); err != nil {
			errorMsg := fmt.Errorf("failed to generate content for page %s: %w", page.ID, err)
			result.Errors = append(result.Errors, errorMsg)
			g.logger.Error("Page generation failed", "page", page.ID, "error", err)
			continue
		}

		result.Pages[page.ID] = pagePtr
		result.TotalWords += pagePtr.WordCount
	}

	result.TotalPages = len(result.Pages)
	options.ProgressTracker.CompleteTask(fmt.Sprintf("Generated %d pages", result.TotalPages))

	g.logger.Info("Wiki generation completed",
		"total_pages", result.TotalPages,
		"total_words", result.TotalWords,
		"errors", len(result.Errors),
		"duration", result.ProcessingTime)

	return result, nil
}

// buildFileTree creates a string representation of the file tree
func (g *WikiGenerator) buildFileTree(files []scanner.FileInfo, basePath string) string {
	var builder strings.Builder

	// Group files by directory
	dirMap := make(map[string][]scanner.FileInfo)

	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if dir == "." {
			dir = ""
		}
		dirMap[dir] = append(dirMap[dir], file)
	}

	// TODO: Improve file tree structure with proper hierarchy and indentation
	for dir, dirFiles := range dirMap {
		if dir != "" {
			builder.WriteString(fmt.Sprintf("%s/\n", dir))
		}
		for _, file := range dirFiles {
			if dir != "" {
				builder.WriteString("  ")
			}
			builder.WriteString(fmt.Sprintf("- %s\n", filepath.Base(file.Path)))
		}
	}

	return builder.String()
}

// findReadmeContent finds and reads README file content
func (g *WikiGenerator) findReadmeContent(files []scanner.FileInfo) string {
	readmeNames := []string{"README.md", "readme.md", "README", "readme", "README.txt"}
	for _, file := range files {
		filename := filepath.Base(file.Path)
		for _, readmeName := range readmeNames {
			if filename == readmeName {
				// Actually read and return README file content
				content, err := os.ReadFile(file.Path)
				if err != nil {
					return "No README file found."
				}
				return string(content)
			}
		}
	}

	return "No README file found."
}

// formatRelevantFiles formats relevant documents for the prompt
func (g *WikiGenerator) formatRelevantFiles(docs []rag.RetrievalResult) string {
	var builder strings.Builder

	for _, doc := range docs {
		builder.WriteString(fmt.Sprintf("\n--- %s ---\n", doc.FilePath))
		builder.WriteString(doc.Content)
		builder.WriteString("\n")
	}

	return builder.String()
}
