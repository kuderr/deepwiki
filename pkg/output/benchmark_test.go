package output

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/generator"
	outputgen "github.com/kuderr/deepwiki/pkg/output/generator"
)

// BenchmarkOutputManager_GenerateMarkdown benchmarks markdown generation
func BenchmarkOutputManager_GenerateMarkdown(b *testing.B) {
	manager := NewOutputManager()
	tempDir := b.TempDir()

	structure, pages := createBenchmarkWikiData(100) // 100 pages

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatMarkdown,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "benchmark-project",
		ProjectPath: "/benchmark/project",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateOutput(structure, pages, options)
		if err != nil {
			b.Fatalf("GenerateOutput failed: %v", err)
		}
	}
}

// BenchmarkOutputManager_GenerateJSON benchmarks JSON generation
func BenchmarkOutputManager_GenerateJSON(b *testing.B) {
	manager := NewOutputManager()
	tempDir := b.TempDir()

	structure, pages := createBenchmarkWikiData(100) // 100 pages

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatJSON,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "benchmark-project",
		ProjectPath: "/benchmark/project",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateOutput(structure, pages, options)
		if err != nil {
			b.Fatalf("GenerateOutput failed: %v", err)
		}
	}
}

// BenchmarkConcurrentProcessor_ProcessPages benchmarks concurrent processing
func BenchmarkConcurrentProcessor_ProcessPages(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}

	concurrencyLevels := []int{1, 2, 4, 8}
	pageCounts := []int{10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		for _, pageCount := range pageCounts {
			b.Run(fmt.Sprintf("Concurrency%d_Pages%d", concurrency, pageCount), func(b *testing.B) {
				cp := NewConcurrentProcessor(logger, concurrency, tracker)

				pages := createBenchmarkPages(pageCount)

				processor := func(page *generator.WikiPage) error {
					// Simulate some work
					time.Sleep(time.Microsecond * 100)
					return nil
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					cp.ProcessPagesParallel(context.Background(), pages, processor)
				}
			})
		}
	}
}

// BenchmarkMemoryEfficientProcessor benchmarks memory management
func BenchmarkMemoryEfficientProcessor(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mp := NewMemoryEfficientProcessor(logger, 100) // 100MB limit

	processor := func() error {
		// Simulate memory allocation
		data := make([]byte, 1024*1024) // 1MB
		_ = data
		return nil
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mp.ProcessWithMemoryManagement(context.Background(), processor)
	}
}

// BenchmarkBatchProcessor benchmarks batch processing
func BenchmarkBatchProcessor(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	batchSizes := []int{5, 10, 20, 50}
	itemCounts := []int{100, 500, 1000}

	for _, batchSize := range batchSizes {
		for _, itemCount := range itemCounts {
			b.Run(fmt.Sprintf("BatchSize%d_Items%d", batchSize, itemCount), func(b *testing.B) {
				bp := NewBatchProcessor[string](logger, batchSize, 2)

				items := make([]string, itemCount)
				for i := 0; i < itemCount; i++ {
					items[i] = fmt.Sprintf("item-%d", i)
				}

				processor := func(batch []string) error {
					// Simulate processing
					for range batch {
						time.Sleep(time.Microsecond * 10)
					}
					return nil
				}

				tracker := &generator.NoOpProgressTracker{}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					bp.ProcessInBatches(context.Background(), items, processor, tracker)
				}
			})
		}
	}
}

// BenchmarkProgressTracker benchmarks progress tracking
func BenchmarkProgressTracker(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	b.Run("Enhanced", func(b *testing.B) {
		tracker := NewEnhancedProgressTracker(logger)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			tracker.StartTask("benchmark", 100)
			for j := 0; j < 100; j++ {
				tracker.UpdateProgress(j, "processing")
			}
			tracker.CompleteTask("done")
		}
	})

	b.Run("Console", func(b *testing.B) {
		tracker := generator.NewConsoleProgressTracker(logger)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			tracker.StartTask("benchmark", 100)
			for j := 0; j < 100; j++ {
				tracker.UpdateProgress(j, "processing")
			}
			tracker.CompleteTask("done")
		}
	})
}

// Helper functions for benchmarks

func createBenchmarkWikiData(pageCount int) (*generator.WikiStructure, map[string]*generator.WikiPage) {
	structure := &generator.WikiStructure{
		ID:          "benchmark-wiki",
		Title:       "Benchmark Wiki",
		Description: "A wiki structure for benchmarking purposes",
		Version:     "1.0.0",
		Language:    "en",
		ProjectPath: "/benchmark/project",
		CreatedAt:   time.Now(),
	}

	pages := make(map[string]*generator.WikiPage, pageCount)
	for i := 0; i < pageCount; i++ {
		pageID := fmt.Sprintf("page-%d", i)
		pages[pageID] = &generator.WikiPage{
			ID:          pageID,
			Title:       fmt.Sprintf("Benchmark Page %d", i),
			Description: fmt.Sprintf("Description for benchmark page %d", i),
			Content:     generateBenchmarkContent(i),
			Importance:  getImportanceLevel(i),
			WordCount:   500 + (i * 10), // Vary word count
			SourceFiles: 1 + (i % 5),    // 1-5 source files
			FilePaths:   []string{fmt.Sprintf("src/file%d.go", i)},
			CreatedAt:   time.Now(),
		}
	}

	return structure, pages
}

func createBenchmarkPages(count int) []*generator.WikiPage {
	pages := make([]*generator.WikiPage, count)
	for i := 0; i < count; i++ {
		pages[i] = &generator.WikiPage{
			ID:          fmt.Sprintf("bench-page-%d", i),
			Title:       fmt.Sprintf("Benchmark Page %d", i),
			Content:     generateBenchmarkContent(i),
			Importance:  getImportanceLevel(i),
			WordCount:   100 + (i * 5),
			SourceFiles: 1,
			CreatedAt:   time.Now(),
		}
	}
	return pages
}

func generateBenchmarkContent(index int) string {
	return fmt.Sprintf(`# Benchmark Page %d

## Overview

This is benchmark page number %d, created for performance testing purposes.

## Features

- Feature A: High-performance processing
- Feature B: Efficient memory usage
- Feature C: Concurrent operations

## Code Example

'''go
func benchmarkFunction%d() {
    // Implementation for benchmark %d
    for i := 0; i < 1000; i++ {
        processData(i)
    }
}
'''

## Mermaid Diagram

'''mermaid
graph TD
    A[Start] --> B{Decision %d}
    B -->|Yes| C[Process]
    B -->|No| D[Skip]
    C --> E[End]
    D --> E
'''

This content is generated for benchmarking the documentation system.
Page %d contains realistic content structure similar to actual documentation.
`, index, index, index, index, index, index)
}

func getImportanceLevel(index int) string {
	switch index % 3 {
	case 0:
		return "high"
	case 1:
		return "medium"
	default:
		return "low"
	}
}
