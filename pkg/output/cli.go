package output

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	outputgen "github.com/deepwiki-cli/deepwiki-cli/pkg/output/generator"
)

// CLIManager manages the command-line interface experience
type CLIManager struct {
	logger          *slog.Logger
	verboseMode     bool
	quietMode       bool
	dryRunMode      bool
	progressTracker *EnhancedProgressTracker
	multiTracker    *MultiTaskTracker
	startTime       time.Time
	stats           *OperationStats
}

// OperationStats tracks statistics for the current operation
type OperationStats struct {
	FilesScanned      int
	FilesProcessed    int
	ChunksGenerated   int
	EmbeddingsCreated int
	PagesGenerated    int
	TokensUsed        int
	APICallsMade      int
	ErrorsEncountered int
	StartTime         time.Time
	EndTime           time.Time
	Phases            map[string]PhaseStats
}

// PhaseStats tracks statistics for a specific phase
type PhaseStats struct {
	Name           string
	StartTime      time.Time
	EndTime        time.Time
	ItemsProcessed int
	ErrorsCount    int
}

// NewCLIManager creates a new CLI manager
func NewCLIManager(logger *slog.Logger, verbose, quiet, dryRun bool) *CLIManager {
	return &CLIManager{
		logger:          logger,
		verboseMode:     verbose,
		quietMode:       quiet,
		dryRunMode:      dryRun,
		progressTracker: NewEnhancedProgressTracker(logger),
		multiTracker:    NewMultiTaskTracker(logger),
		startTime:       time.Now(),
		stats: &OperationStats{
			StartTime: time.Now(),
			Phases:    make(map[string]PhaseStats),
		},
	}
}

// StartOperation starts a new operation with welcome banner
func (c *CLIManager) StartOperation(projectName, outputDir string) {
	if c.quietMode {
		return
	}

	c.printBanner()

	if c.dryRunMode {
		fmt.Println("🧪 DRY RUN MODE - No files will be generated")
		fmt.Println()
	}

	fmt.Printf("📁 Project: %s\n", projectName)
	fmt.Printf("📂 Output: %s\n", outputDir)
	fmt.Println()
}

// StartPhase starts a new phase of the operation
func (c *CLIManager) StartPhase(phaseName string, description string, total int) {
	if !c.quietMode {
		fmt.Printf("🔄 %s: %s\n", phaseName, description)
	}

	c.stats.Phases[phaseName] = PhaseStats{
		Name:      phaseName,
		StartTime: time.Now(),
	}

	if total > 0 {
		c.progressTracker.StartTask(phaseName, total)
	}
}

// UpdatePhase updates the progress of the current phase
func (c *CLIManager) UpdatePhase(current int, message string) {
	if c.verboseMode && !c.quietMode && message != "" {
		// In verbose mode, show detailed messages
		c.logger.Info("Phase progress", "current", current, "message", message)
	}

	c.progressTracker.UpdateProgress(current, message)
}

// CompletePhase completes the current phase
func (c *CLIManager) CompletePhase(phaseName string, itemsProcessed int, errors int) {
	phase := c.stats.Phases[phaseName]
	phase.EndTime = time.Now()
	phase.ItemsProcessed = itemsProcessed
	phase.ErrorsCount = errors
	c.stats.Phases[phaseName] = phase

	var message string
	if errors > 0 {
		message = fmt.Sprintf("%d items processed, %d errors", itemsProcessed, errors)
	} else {
		message = fmt.Sprintf("%d items processed", itemsProcessed)
	}

	c.progressTracker.CompleteTask(message)
}

// ReportError reports an error that occurred during processing
func (c *CLIManager) ReportError(phase string, err error, context string) {
	c.stats.ErrorsEncountered++

	if phase != "" {
		if phaseStats, exists := c.stats.Phases[phase]; exists {
			phaseStats.ErrorsCount++
			c.stats.Phases[phase] = phaseStats
		}
	}

	c.progressTracker.SetError(err)

	if c.verboseMode {
		c.logger.Error("Operation error",
			"phase", phase,
			"error", err.Error(),
			"context", context)
	}
}

// CompleteOperation completes the operation and shows summary
func (c *CLIManager) CompleteOperation(result *outputgen.OutputResult, errors []error) {
	c.stats.EndTime = time.Now()

	if c.dryRunMode {
		c.showDryRunSummary(result)
		return
	}

	if len(errors) > 0 {
		c.showErrorSummary(errors)
	}

	c.showSuccessSummary(result)
}

// ShowVerboseStats shows detailed statistics in verbose mode
func (c *CLIManager) ShowVerboseStats() {
	if !c.verboseMode || c.quietMode {
		return
	}

	fmt.Println("\n📊 Detailed Statistics:")

	// Show phase breakdown
	for phaseName, phase := range c.stats.Phases {
		duration := "unknown"
		if !phase.EndTime.IsZero() {
			duration = phase.EndTime.Sub(phase.StartTime).Round(time.Second).String()
		}

		fmt.Printf("  📋 %s: %d items, %d errors, %s\n",
			phaseName, phase.ItemsProcessed, phase.ErrorsCount, duration)
	}

	// Show resource usage
	fmt.Printf("  🔢 API calls: %d\n", c.stats.APICallsMade)
	fmt.Printf("  📝 Tokens used: %d\n", c.stats.TokensUsed)
	fmt.Printf("  🧮 Embeddings: %d\n", c.stats.EmbeddingsCreated)
	fmt.Printf("  ⚡ Chunks: %d\n", c.stats.ChunksGenerated)
}

// UpdateStats updates various statistics
func (c *CLIManager) UpdateStats(field string, value int) {
	switch field {
	case "files_scanned":
		c.stats.FilesScanned = value
	case "files_processed":
		c.stats.FilesProcessed = value
	case "chunks_generated":
		c.stats.ChunksGenerated = value
	case "embeddings_created":
		c.stats.EmbeddingsCreated = value
	case "pages_generated":
		c.stats.PagesGenerated = value
	case "tokens_used":
		c.stats.TokensUsed += value
	case "api_calls":
		c.stats.APICallsMade += value
	}
}

// Helper methods

func (c *CLIManager) printBanner() {
	if c.quietMode {
		return
	}

	banner := `
╔═══════════════════════════════════════╗
║           🤖 DeepWiki CLI             ║
║     AI-Powered Documentation          ║
╚═══════════════════════════════════════╝
`
	fmt.Println(banner)
}

func (c *CLIManager) showSuccessSummary(result *outputgen.OutputResult) {
	if c.quietMode {
		fmt.Printf("Generated %d files in %s\n", result.TotalFiles, result.OutputDir)
		return
	}

	fmt.Println("\n🎉 Generation completed successfully!")
	fmt.Printf("📁 Output directory: %s\n", result.OutputDir)
	fmt.Printf("📄 Files generated: %d\n", result.TotalFiles)
	fmt.Printf("💾 Total size: %s\n", formatBytes(result.TotalSize))
	fmt.Printf("⏱️  Processing time: %v\n", result.ProcessingTime.Round(time.Second))

	// Show key files
	if len(result.FilesGenerated) > 0 {
		fmt.Println("\n📋 Key files:")
		keyFiles := []string{"index.md", "index.json", "wiki.json", "wiki-structure.json"}

		for _, keyFile := range keyFiles {
			for _, generated := range result.FilesGenerated {
				if strings.HasSuffix(generated, keyFile) {
					fmt.Printf("  • %s\n", generated)
					break
				}
			}
		}
	}

	if c.verboseMode {
		c.ShowVerboseStats()
	}
}

func (c *CLIManager) showErrorSummary(errors []error) {
	if c.quietMode {
		fmt.Printf("Completed with %d errors\n", len(errors))
		return
	}

	fmt.Printf("\n⚠️  Completed with %d errors:\n", len(errors))

	for i, err := range errors {
		if i < 5 { // Show first 5 errors
			fmt.Printf("  %d. %v\n", i+1, err)
		} else if i == 5 {
			fmt.Printf("  ... and %d more errors\n", len(errors)-5)
			break
		}
	}

	if c.verboseMode {
		fmt.Println("\nFor detailed error information, check the logs.")
	}
}

func (c *CLIManager) showDryRunSummary(result *outputgen.OutputResult) {
	fmt.Println("\n🧪 Dry run completed!")
	fmt.Printf("📁 Would create output in: %s\n", result.OutputDir)
	fmt.Printf("📄 Would generate: %d files\n", result.TotalFiles)
	fmt.Printf("⏱️  Analysis time: %v\n", result.ProcessingTime.Round(time.Second))

	if len(result.Errors) > 0 {
		fmt.Printf("⚠️  Would encounter %d errors\n", len(result.Errors))
	}

	fmt.Println("\nTo actually generate files, run without --dry-run flag.")
}

// formatBytes formats byte size in human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div),
		"KMGTPE"[exp])
}

// GetOperationSummary returns a summary of the operation for external use
func (c *CLIManager) GetOperationSummary() OperationSummary {
	totalDuration := time.Since(c.startTime)
	if !c.stats.EndTime.IsZero() {
		totalDuration = c.stats.EndTime.Sub(c.stats.StartTime)
	}

	return OperationSummary{
		TotalDuration:     totalDuration,
		FilesScanned:      c.stats.FilesScanned,
		FilesProcessed:    c.stats.FilesProcessed,
		PagesGenerated:    c.stats.PagesGenerated,
		ErrorsEncountered: c.stats.ErrorsEncountered,
		TokensUsed:        c.stats.TokensUsed,
		APICallsMade:      c.stats.APICallsMade,
		Phases:            len(c.stats.Phases),
	}
}

// OperationSummary provides a summary of the operation
type OperationSummary struct {
	TotalDuration     time.Duration
	FilesScanned      int
	FilesProcessed    int
	PagesGenerated    int
	ErrorsEncountered int
	TokensUsed        int
	APICallsMade      int
	Phases            int
}
