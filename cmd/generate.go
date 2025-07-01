package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kuderr/deepwiki/internal/config"
	"github.com/kuderr/deepwiki/internal/logging"
	"github.com/kuderr/deepwiki/pkg/embeddings"
	"github.com/kuderr/deepwiki/pkg/generator"
	"github.com/kuderr/deepwiki/pkg/openai"
	"github.com/kuderr/deepwiki/pkg/output"
	outputgen "github.com/kuderr/deepwiki/pkg/output/generator"
	"github.com/kuderr/deepwiki/pkg/processor"
	"github.com/kuderr/deepwiki/pkg/rag"
	"github.com/kuderr/deepwiki/pkg/scanner"
	"github.com/kuderr/deepwiki/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// Command flags
	projectPath  string
	outputDir    string
	format       string
	language     string
	openaiKey    string
	model        string
	excludeDirs  string
	excludeFiles string
	chunkSize    int
	configFile   string
	verbose      bool
	dryRun       bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [directory]",
	Short: "Generate documentation for a local directory",
	Long: `Generate comprehensive documentation for a local directory using AI.

This command scans the specified directory (or current directory if not specified),
analyzes the codebase, and generates structured documentation including:
- Project overview and architecture
- Code analysis and explanations
- API documentation
- Setup and deployment guides

The generated documentation can be output in multiple formats (markdown, json)
and supports multiple languages.

Examples:
  deepwiki-cli generate
  deepwiki-cli generate /path/to/project
  deepwiki-cli generate --output-dir ./docs
  deepwiki-cli generate --format json --language ja`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with CLI flags
	overrideConfigWithFlags(cfg, cmd)

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Close()

	// Set global logger
	logging.SetGlobalLogger(logger)

	// Create context for the operation
	ctx := context.Background()

	// Create a component-specific logger
	genLogger := logger.WithComponent("generator")

	// Determine the project path
	if len(args) > 0 {
		projectPath = args[0]
	} else if projectPath == "" {
		var err error
		projectPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	projectPath = absPath

	// Validate project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", projectPath)
	}

	// Validate output directory
	if cfg.Output.Directory != "" {
		if err := os.MkdirAll(cfg.Output.Directory, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Validate OpenAI API key
	if cfg.OpenAI.APIKey == "" {
		genLogger.ErrorContext(ctx, "OpenAI API key is required")
		return fmt.Errorf("OpenAI API key is required. Set --openai-key flag or OPENAI_API_KEY environment variable")
	}

	genLogger.InfoContext(ctx, "starting documentation generation",
		slog.String("project_path", projectPath),
		slog.String("output_dir", cfg.Output.Directory),
		slog.String("format", cfg.Output.Format),
		slog.String("language", cfg.Output.Language),
	)

	if verbose {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  Project Path: %s\n", projectPath)
		fmt.Printf("  Output Dir: %s\n", cfg.Output.Directory)
		fmt.Printf("  Format: %s\n", cfg.Output.Format)
		fmt.Printf("  Language: %s\n", cfg.Output.Language)
		fmt.Printf("  Model: %s\n", cfg.OpenAI.Model)
		fmt.Printf("  Embedding Model: %s\n", cfg.OpenAI.EmbeddingModel)
		fmt.Printf("  Chunk Size: %d\n", cfg.Processing.ChunkSize)
		fmt.Printf("  Max Tokens: %d\n", cfg.OpenAI.MaxTokens)
		fmt.Printf("  Temperature: %.1f\n", cfg.OpenAI.Temperature)
		if len(cfg.Filters.ExcludeDirs) > 0 {
			fmt.Printf("  Exclude Dirs: %v\n", cfg.Filters.ExcludeDirs)
		}
		if len(cfg.Filters.ExcludeFiles) > 0 {
			fmt.Printf("  Exclude Files: %v\n", cfg.Filters.ExcludeFiles)
		}
		fmt.Printf("\n")
	}

	if dryRun {
		fmt.Println("Dry run mode - no documentation will be generated")
		return nil
	}

	// Start documentation generation
	fmt.Printf("🚀 Starting documentation generation for: %s\n", projectPath)
	fmt.Printf("📝 Output will be saved to: %s\n", cfg.Output.Directory)
	fmt.Println()

	// Step 1: Scan directory
	fmt.Println("📁 Scanning directory...")
	genLogger.InfoContext(ctx, "starting directory scan", slog.String("path", projectPath))

	scanOptions := &scanner.ScanOptions{
		IncludeExtensions: cfg.Filters.IncludeExtensions,
		ExcludeDirs:       cfg.Filters.ExcludeDirs,
		ExcludeFiles:      cfg.Filters.ExcludeFiles,
		FollowSymlinks:    false,
		MaxDepth:          0, // unlimited
		MaxFiles:          cfg.Processing.MaxFiles,
		AnalyzeContent:    true,
		MaxFileSize:       1024 * 1024, // 1MB
		SkipBinaryFiles:   true,
		Concurrent:        true,
		MaxWorkers:        4,
	}

	fileScanner := scanner.NewScanner(scanOptions)
	scanResult, err := fileScanner.ScanDirectory(projectPath)
	if err != nil {
		genLogger.LogError(ctx, "directory scan failed", err, slog.String("path", projectPath))
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Log scan results
	logger.LogScanResult(ctx, scanResult.TotalFiles, scanResult.FilteredFiles, scanResult.ScanTime)

	// Display scan results
	fmt.Printf("✅ Directory scan completed in %v\n", scanResult.ScanTime.Round(time.Millisecond))
	fmt.Printf("   • Found %d files in %d directories\n", scanResult.TotalFiles, scanResult.TotalDirs)
	fmt.Printf("   • Filtered to %d relevant files\n", scanResult.FilteredFiles)

	if len(scanResult.Errors) > 0 {
		fmt.Printf("   • %d errors occurred during scanning\n", len(scanResult.Errors))
		if verbose {
			for _, err := range scanResult.Errors {
				fmt.Printf("     - %s\n", err)
			}
		}
	}

	// Show file breakdown by category
	if verbose {
		categories := make(map[string]int)
		languages := make(map[string]int)

		for _, file := range scanResult.Files {
			categories[file.Category]++
			if file.Language != "" {
				languages[file.Language]++
			}
		}

		if len(categories) > 0 {
			fmt.Printf("\n📊 File categories:\n")
			for category, count := range categories {
				fmt.Printf("   • %s: %d files\n", category, count)
			}
		}

		if len(languages) > 0 {
			fmt.Printf("\n🔧 Programming languages:\n")
			for language, count := range languages {
				fmt.Printf("   • %s: %d files\n", language, count)
			}
		}
	}

	fmt.Println()

	// Initialize CLI manager for progress tracking
	cliManager := output.NewCLIManager(genLogger.Logger, verbose, false, dryRun)
	cliManager.StartOperation(filepath.Base(projectPath), cfg.Output.Directory)

	// Initialize OpenAI client
	openaiConfig := &openai.Config{
		APIKey:         cfg.OpenAI.APIKey,
		Model:          cfg.OpenAI.Model,
		EmbeddingModel: cfg.OpenAI.EmbeddingModel,
		MaxTokens:      cfg.OpenAI.MaxTokens,
		Temperature:    float64(cfg.OpenAI.Temperature),
		RequestTimeout: 3 * 60 * time.Second,
		MaxRetries:     5,
		RetryDelay:     1 * time.Second,
		RateLimitRPS:   10,
	}
	openaiClient, err := openai.NewClient(openaiConfig)
	if err != nil {
		genLogger.LogError(ctx, "failed to initialize OpenAI client", err)
		return fmt.Errorf("failed to initialize OpenAI client: %w", err)
	}

	// Phase 2: Text Processing and Chunking
	cliManager.StartPhase("Phase 2", "Processing and chunking files", len(scanResult.Files))
	fmt.Println("📝 Phase 2: Processing and chunking files...")

	processingOptions := &processor.ProcessingOptions{
		ChunkSize:           cfg.Processing.ChunkSize,
		ChunkOverlap:        50,
		MaxChunkWords:       500,
		MinChunkWords:       10,
		MaxChunks:           0,
		Concurrent:          true,
		MaxWorkers:          4,
		CountTokens:         true,
		NormalizeWhitespace: true,
		RemoveComments:      false,
	}

	textProcessor := processor.NewTextProcessor(processingOptions)
	processingResult, err := textProcessor.ProcessFiles(scanResult.Files)
	if err != nil {
		cliManager.ReportError("Phase 2", err, "text processing failed")
		return fmt.Errorf("failed to process files: %w", err)
	}

	cliManager.CompletePhase("Phase 2", len(processingResult.Documents), len(processingResult.Errors))
	fmt.Printf("✅ Phase 2 completed: %d documents processed, %d chunks created\n",
		len(processingResult.Documents), processingResult.TotalChunks)

	// Phase 3: Embedding Generation
	cliManager.StartPhase("Phase 3", "Generating embeddings", processingResult.TotalChunks)
	fmt.Println("🧠 Phase 3: Generating embeddings...")

	embeddingConfig := embeddings.DefaultEmbeddingConfig()
	embeddingConfig.Model = cfg.OpenAI.EmbeddingModel

	embeddingGenerator := embeddings.NewOpenAIEmbeddingGenerator(openaiClient, embeddingConfig)

	// Collect all chunk texts for embedding
	var chunkTexts []string
	var chunkMetadata []map[string]interface{}

	for _, doc := range processingResult.Documents {
		for _, chunk := range doc.Chunks {
			chunkTexts = append(chunkTexts, chunk.Text)
			metadata := map[string]interface{}{
				"document_id": doc.ID,
				"file_path":   doc.FilePath,
				"chunk_id":    chunk.ID,
				"language":    doc.Language,
				"category":    doc.Category,
			}
			chunkMetadata = append(chunkMetadata, metadata)
		}
	}

	embeddingVectors, err := embeddingGenerator.GenerateBatchEmbeddings(chunkTexts)
	if err != nil {
		cliManager.ReportError("Phase 3", err, "embedding generation failed")
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	cliManager.CompletePhase("Phase 3", len(embeddingVectors), 0)
	fmt.Printf("✅ Phase 3 completed: %d embeddings generated\n", len(embeddingVectors))

	// Phase 4: RAG Setup and Document Indexing
	cliManager.StartPhase("Phase 4", "Setting up RAG and indexing documents", len(embeddingVectors))
	fmt.Println("🔍 Phase 4: Setting up RAG and indexing documents...")

	// Create vector database
	vectorDB, err := embeddings.NewBoltVectorDB(embeddingConfig)
	if err != nil {
		cliManager.ReportError("Phase 4", err, "failed to create vector database")
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	// Create document embeddings and store them
	for docIndex, doc := range processingResult.Documents {
		docEmbeddings := make([]embeddings.EmbeddingVector, 0, len(doc.Chunks))

		for chunkIndex, chunk := range doc.Chunks {
			globalIndex := docIndex*len(doc.Chunks) + chunkIndex
			if globalIndex < len(embeddingVectors) && embeddingVectors[globalIndex] != nil {
				embVector := embeddings.EmbeddingVector{
					ID:        chunk.ID,
					Vector:    embeddingVectors[globalIndex],
					Content:   chunk.Text,
					Dimension: len(embeddingVectors[globalIndex]),
					Metadata: map[string]string{
						"document_id": doc.ID,
						"file_path":   doc.FilePath,
						"language":    doc.Language,
						"category":    doc.Category,
					},
					CreatedAt: time.Now(),
				}
				docEmbeddings = append(docEmbeddings, embVector)
			}
		}

		if len(docEmbeddings) > 0 {
			docEmbedding := &embeddings.DocumentEmbedding{
				DocumentID:  doc.ID,
				FilePath:    doc.FilePath,
				Language:    doc.Language,
				Category:    doc.Category,
				ChunkCount:  len(docEmbeddings),
				Embeddings:  docEmbeddings,
				ProcessedAt: time.Now(),
			}

			err := vectorDB.Store(docEmbedding)
			if err != nil {
				genLogger.LogError(ctx, "failed to store document embedding", err,
					slog.String("document_id", doc.ID))
				continue
			}
		}
	}

	// Initialize RAG retriever
	embeddingService := embeddings.NewEmbeddingService(embeddingGenerator, vectorDB, embeddingConfig)
	ragRetriever := rag.NewDocumentRetriever(
		embeddingService,
		vectorDB,
		embeddingGenerator,
		processingResult.Documents,
		nil,
	)

	cliManager.CompletePhase("Phase 4", len(embeddingVectors), 0)
	fmt.Printf("✅ Phase 4 completed: %d documents indexed in vector database\n", len(embeddingVectors))

	// Phase 5: Wiki Structure Generation
	cliManager.StartPhase("Phase 5", "Generating wiki structure", 1)
	fmt.Println("🏗️  Phase 5: Generating wiki structure...")

	wikiGenerator := generator.NewWikiGenerator(openaiClient, ragRetriever, genLogger.Logger)

	progressTracker := generator.NewConsoleProgressTracker(genLogger.Logger)

	generationOptions := generator.GenerationOptions{
		ProjectName:     filepath.Base(projectPath),
		ProjectPath:     projectPath,
		Language:        cfg.Output.Language,
		OutputFormat:    cfg.Output.Format,
		ProgressTracker: progressTracker,
	}

	generationResult, err := wikiGenerator.GenerateWiki(ctx, scanResult.Files, generationOptions)
	if err != nil {
		cliManager.ReportError("Phase 5", err, "wiki generation failed")
		return fmt.Errorf("failed to generate wiki: %w", err)
	}

	cliManager.CompletePhase("Phase 5", generationResult.TotalPages, len(generationResult.Errors))
	fmt.Printf("✅ Phase 5 completed: Wiki structure with %d pages generated\n", generationResult.TotalPages)

	// Phase 6: Content Generation and Output
	cliManager.StartPhase("Phase 6", "Generating final output", generationResult.TotalPages+1)
	fmt.Println("📄 Phase 6: Generating final output...")

	// Create output manager
	outputManager := output.NewOutputManager()

	// Prepare output options
	outputOptions := outputgen.OutputOptions{
		Directory:   cfg.Output.Directory,
		Format:      outputgen.OutputFormat(cfg.Output.Format),
		ProjectName: generationOptions.ProjectName,
		Language:    cfg.Output.Language,
	}

	// Generate output files
	outputResult, err := outputManager.GenerateOutput(generationResult.Structure, generationResult.Pages, outputOptions)
	if err != nil {
		cliManager.ReportError("Phase 6", err, "output generation failed")
		return fmt.Errorf("failed to generate output: %w", err)
	}

	cliManager.CompletePhase("Phase 6", outputResult.TotalFiles, len(outputResult.Errors))
	cliManager.CompleteOperation(outputResult, outputResult.Errors)

	// Show final summary
	fmt.Printf("\n🎉 Documentation generation completed!\n")
	fmt.Printf("📁 Output directory: %s\n", cfg.Output.Directory)
	fmt.Printf("📄 Files generated: %d\n", outputResult.TotalFiles)
	fmt.Printf("📝 Total pages: %d\n", generationResult.TotalPages)
	fmt.Printf("🔤 Total words: %d\n", generationResult.TotalWords)
	fmt.Printf("⏱️  Total processing time: %v\n", time.Since(time.Now().Add(-generationResult.ProcessingTime)))

	if len(outputResult.Errors) > 0 {
		fmt.Printf("\n⚠️  %d errors occurred during generation\n", len(outputResult.Errors))
		if verbose {
			for i, err := range outputResult.Errors {
				fmt.Printf("  %d. %v\n", i+1, err)
			}
		}
	}

	return nil
}

// overrideConfigWithFlags overrides configuration values with CLI flags when provided
func overrideConfigWithFlags(cfg *config.Config, cmd *cobra.Command) {
	if outputDir != "" {
		cfg.Output.Directory = outputDir
	}
	if format != "" {
		cfg.Output.Format = format
	}
	if language != "" {
		if parsedLang, err := types.ParseLanguageWithCode(language); err == nil {
			cfg.Output.Language = parsedLang
		} else {
			fmt.Printf("Warning: Invalid language flag '%s', using default. %s\n", language, err.Error())
		}
	}
	if openaiKey != "" {
		cfg.OpenAI.APIKey = openaiKey
	}
	if model != "" {
		cfg.OpenAI.Model = model
	}
	if chunkSize > 0 {
		cfg.Processing.ChunkSize = chunkSize
	}

	// Handle comma-separated exclude options
	if excludeDirs != "" {
		for _, dir := range strings.Split(excludeDirs, ",") {
			cfg.Filters.ExcludeDirs = append(cfg.Filters.ExcludeDirs, strings.TrimSpace(dir))
		}
	}
	if excludeFiles != "" {
		for _, file := range strings.Split(excludeFiles, ",") {
			cfg.Filters.ExcludeFiles = append(cfg.Filters.ExcludeFiles, strings.TrimSpace(file))
		}
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Persistent flags
	generateCmd.Flags().
		StringVarP(&projectPath, "path", "p", "", "Path to the project directory (default: current directory)")
	generateCmd.Flags().
		StringVarP(&outputDir, "output-dir", "o", "./docs", "Output directory for generated documentation")
	generateCmd.Flags().
		StringVarP(&format, "format", "f", "markdown", "Output format: markdown, json, docusaurus2, docusaurus3, simple-docusaurus2, simple-docusaurus3")
	generateCmd.Flags().
		StringVarP(&language, "language", "l", "en", "Language for generation: en, ru, ja, zh, es, kr, vi")
	generateCmd.Flags().StringVar(&openaiKey, "openai-key", "", "OpenAI API key (or use OPENAI_API_KEY env var)")
	generateCmd.Flags().StringVarP(&model, "model", "m", "gpt-4o", "OpenAI model to use")
	generateCmd.Flags().StringVar(&excludeDirs, "exclude-dirs", "", "Comma-separated list of directories to exclude")
	generateCmd.Flags().StringVar(&excludeFiles, "exclude-files", "", "Comma-separated patterns for files to exclude")
	generateCmd.Flags().IntVar(&chunkSize, "chunk-size", 350, "Text chunk size for embeddings")
	generateCmd.Flags().StringVar(&configFile, "config", "", "Configuration file path")
	generateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	generateCmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "Show what would be done without actually generating documentation")

	// Mark required flags
	// Note: OpenAI key will be checked in the run function to allow env var
}
