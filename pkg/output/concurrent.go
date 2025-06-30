package output

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// ConcurrentProcessor provides concurrent processing capabilities
type ConcurrentProcessor struct {
	logger          *slog.Logger
	maxConcurrency  int
	progressTracker generator.ProgressTracker
}

// NewConcurrentProcessor creates a new concurrent processor
func NewConcurrentProcessor(
	logger *slog.Logger,
	maxConcurrency int,
	tracker generator.ProgressTracker,
) *ConcurrentProcessor {
	if maxConcurrency <= 0 {
		maxConcurrency = runtime.NumCPU()
	}

	return &ConcurrentProcessor{
		logger:          logger.With("component", "concurrent"),
		maxConcurrency:  maxConcurrency,
		progressTracker: tracker,
	}
}

// ProcessPagesParallel processes multiple wiki pages concurrently
func (cp *ConcurrentProcessor) ProcessPagesParallel(
	ctx context.Context,
	pages []*generator.WikiPage,
	processor func(*generator.WikiPage) error,
) error {
	if len(pages) == 0 {
		return nil
	}

	// Create worker pool
	jobs := make(chan *generator.WikiPage, len(pages))
	results := make(chan error, len(pages))

	// Start workers
	numWorkers := cp.maxConcurrency
	if numWorkers > len(pages) {
		numWorkers = len(pages)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			cp.worker(ctx, workerID, jobs, results, processor)
		}(i)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for _, page := range pages {
			select {
			case jobs <- page:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	processed := 0
	for err := range results {
		processed++
		if cp.progressTracker != nil {
			cp.progressTracker.UpdateProgress(processed, fmt.Sprintf("Processed %d/%d pages", processed, len(pages)))
		}

		if err != nil {
			errors = append(errors, err)
			cp.logger.ErrorContext(ctx, "page processing error", "error", err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to process %d pages: %v", len(errors), errors)
	}

	return nil
}

// ProcessBatch processes items in batches to optimize memory usage
type BatchProcessor[T any] struct {
	logger      *slog.Logger
	batchSize   int
	concurrency int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor[T any](logger *slog.Logger, batchSize, concurrency int) *BatchProcessor[T] {
	if batchSize <= 0 {
		batchSize = 10
	}
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	return &BatchProcessor[T]{
		logger:      logger.With("component", "batch"),
		batchSize:   batchSize,
		concurrency: concurrency,
	}
}

// ProcessInBatches processes items in batches
func (bp *BatchProcessor[T]) ProcessInBatches(
	ctx context.Context,
	items []T,
	processor func([]T) error,
	progressTracker generator.ProgressTracker,
) error {
	if len(items) == 0 {
		return nil
	}

	// Create batches
	batches := make([][]T, 0)
	for i := 0; i < len(items); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	bp.logger.InfoContext(ctx, "processing in batches",
		"total_items", len(items),
		"batch_size", bp.batchSize,
		"num_batches", len(batches),
		"concurrency", bp.concurrency)

	// Process batches concurrently
	jobs := make(chan []T, len(batches))
	results := make(chan error, len(batches))

	// Start workers
	numWorkers := bp.concurrency
	if numWorkers > len(batches) {
		numWorkers = len(batches)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for batch := range jobs {
				select {
				case <-ctx.Done():
					results <- ctx.Err()
					return
				default:
					if err := processor(batch); err != nil {
						results <- fmt.Errorf("worker %d failed: %w", workerID, err)
					} else {
						results <- nil
					}
				}
			}
		}(i)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for _, batch := range batches {
			select {
			case jobs <- batch:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	processed := 0
	processedItems := 0

	for err := range results {
		processed++
		processedItems += bp.batchSize
		if processedItems > len(items) {
			processedItems = len(items)
		}

		if progressTracker != nil {
			progressTracker.UpdateProgress(
				processedItems,
				fmt.Sprintf("Processed %d/%d items", processedItems, len(items)),
			)
		}

		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch processing failed: %v", errors)
	}

	return nil
}

// MemoryEfficientProcessor processes items with memory constraints
type MemoryEfficientProcessor struct {
	logger        *slog.Logger
	maxMemoryMB   int
	checkInterval time.Duration
	forceGC       bool
	stats         *MemoryStats
	mu            sync.RWMutex
}

// MemoryStats tracks memory usage statistics
type MemoryStats struct {
	StartMemoryMB   float64
	CurrentMemoryMB float64
	PeakMemoryMB    float64
	GCCount         int
	LastGCTime      time.Time
}

// NewMemoryEfficientProcessor creates a new memory-efficient processor
func NewMemoryEfficientProcessor(logger *slog.Logger, maxMemoryMB int) *MemoryEfficientProcessor {
	if maxMemoryMB <= 0 {
		maxMemoryMB = 512 // Default 512MB limit
	}

	processor := &MemoryEfficientProcessor{
		logger:        logger.With("component", "memory"),
		maxMemoryMB:   maxMemoryMB,
		checkInterval: time.Second,
		forceGC:       true,
		stats:         &MemoryStats{},
	}

	// Initialize memory stats
	processor.updateMemoryStats()
	processor.stats.StartMemoryMB = processor.stats.CurrentMemoryMB

	return processor
}

// ProcessWithMemoryManagement processes items with memory management
func (mp *MemoryEfficientProcessor) ProcessWithMemoryManagement(ctx context.Context, processor func() error) error {
	// Start memory monitoring
	stopMonitoring := make(chan bool)
	go mp.monitorMemory(ctx, stopMonitoring)
	defer func() {
		stopMonitoring <- true
	}()

	// Check memory before processing
	if err := mp.checkMemoryLimit(); err != nil {
		return fmt.Errorf("memory check failed: %w", err)
	}

	// Process with memory checks
	err := processor()

	// Force GC after processing
	if mp.forceGC {
		mp.runGC()
	}

	return err
}

// GetMemoryStats returns current memory statistics
func (mp *MemoryEfficientProcessor) GetMemoryStats() MemoryStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return *mp.stats
}

// SetMemoryLimit sets the memory limit
func (mp *MemoryEfficientProcessor) SetMemoryLimit(limitMB int) {
	mp.maxMemoryMB = limitMB
}

// Private methods

func (cp *ConcurrentProcessor) worker(
	ctx context.Context,
	workerID int,
	jobs <-chan *generator.WikiPage,
	results chan<- error,
	processor func(*generator.WikiPage) error,
) {
	for page := range jobs {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			start := time.Now()
			err := processor(page)
			duration := time.Since(start)

			if err != nil {
				cp.logger.ErrorContext(ctx, "worker processing failed",
					"worker_id", workerID,
					"page_id", page.ID,
					"duration", duration,
					"error", err)
			} else {
				cp.logger.DebugContext(ctx, "worker processing completed",
					"worker_id", workerID,
					"page_id", page.ID,
					"duration", duration)
			}

			results <- err
		}
	}
}

func (mp *MemoryEfficientProcessor) monitorMemory(ctx context.Context, stop <-chan bool) {
	ticker := time.NewTicker(mp.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case <-ticker.C:
			mp.updateMemoryStats()

			if mp.stats.CurrentMemoryMB > float64(mp.maxMemoryMB) {
				mp.logger.WarnContext(ctx, "memory usage high",
					"current_mb", mp.stats.CurrentMemoryMB,
					"limit_mb", mp.maxMemoryMB)

				if mp.forceGC {
					mp.runGC()
				}
			}
		}
	}
}

func (mp *MemoryEfficientProcessor) updateMemoryStats() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mp.stats.CurrentMemoryMB = float64(m.Alloc) / 1024 / 1024
	if mp.stats.CurrentMemoryMB > mp.stats.PeakMemoryMB {
		mp.stats.PeakMemoryMB = mp.stats.CurrentMemoryMB
	}
}

func (mp *MemoryEfficientProcessor) checkMemoryLimit() error {
	mp.updateMemoryStats()

	if mp.stats.CurrentMemoryMB > float64(mp.maxMemoryMB) {
		return fmt.Errorf("memory usage (%.1f MB) exceeds limit (%d MB)",
			mp.stats.CurrentMemoryMB, mp.maxMemoryMB)
	}

	return nil
}

func (mp *MemoryEfficientProcessor) runGC() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	runtime.GC()
	mp.stats.GCCount++
	mp.stats.LastGCTime = time.Now()

	mp.logger.Debug("forced garbage collection",
		"gc_count", mp.stats.GCCount,
		"memory_before_mb", mp.stats.CurrentMemoryMB)
}
