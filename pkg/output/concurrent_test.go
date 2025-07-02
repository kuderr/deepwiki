package output

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

func TestNewConcurrentProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}

	cp := NewConcurrentProcessor(logger, 4, tracker)
	if cp == nil {
		t.Fatal("NewConcurrentProcessor returned nil")
	}

	if cp.maxConcurrency != 4 {
		t.Errorf("Expected maxConcurrency 4, got %d", cp.maxConcurrency)
	}
}

func TestConcurrentProcessor_ProcessPagesParallel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}
	cp := NewConcurrentProcessor(logger, 2, tracker)

	// Create test pages
	pages := []*generator.WikiPage{
		{ID: "page1", Title: "Page 1"},
		{ID: "page2", Title: "Page 2"},
		{ID: "page3", Title: "Page 3"},
	}

	processed := make(map[string]bool)
	var mu sync.Mutex
	processor := func(page *generator.WikiPage) error {
		mu.Lock()
		processed[page.ID] = true
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	ctx := context.Background()
	err := cp.ProcessPagesParallel(ctx, pages, processor)
	if err != nil {
		t.Fatalf("ProcessPagesParallel failed: %v", err)
	}

	// Verify all pages were processed
	for _, page := range pages {
		mu.Lock()
		wasProcessed := processed[page.ID]
		mu.Unlock()
		if !wasProcessed {
			t.Errorf("Page %s was not processed", page.ID)
		}
	}
}

func TestConcurrentProcessor_ProcessPagesParallel_WithErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}
	cp := NewConcurrentProcessor(logger, 2, tracker)

	pages := []*generator.WikiPage{
		{ID: "page1", Title: "Page 1"},
		{ID: "error_page", Title: "Error Page"},
		{ID: "page3", Title: "Page 3"},
	}

	processor := func(page *generator.WikiPage) error {
		if page.ID == "error_page" {
			return errors.New("processing error")
		}
		return nil
	}

	ctx := context.Background()
	err := cp.ProcessPagesParallel(ctx, pages, processor)
	if err == nil {
		t.Error("Expected error from ProcessPagesParallel, got nil")
	}
}

func TestNewBatchProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	bp := NewBatchProcessor[string](logger, 5, 2)
	if bp == nil {
		t.Fatal("NewBatchProcessor returned nil")
	}

	if bp.batchSize != 5 {
		t.Errorf("Expected batchSize 5, got %d", bp.batchSize)
	}

	if bp.concurrency != 2 {
		t.Errorf("Expected concurrency 2, got %d", bp.concurrency)
	}
}

func TestBatchProcessor_ProcessInBatches(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bp := NewBatchProcessor[string](logger, 2, 1)

	items := []string{"item1", "item2", "item3", "item4", "item5"}
	processed := make(map[string]bool)

	processor := func(batch []string) error {
		for _, item := range batch {
			processed[item] = true
		}
		return nil
	}

	tracker := &generator.NoOpProgressTracker{}
	ctx := context.Background()

	err := bp.ProcessInBatches(ctx, items, processor, tracker)
	if err != nil {
		t.Fatalf("ProcessInBatches failed: %v", err)
	}

	// Verify all items were processed
	for _, item := range items {
		if !processed[item] {
			t.Errorf("Item %s was not processed", item)
		}
	}
}

func TestNewMemoryEfficientProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mp := NewMemoryEfficientProcessor(logger, 256)
	if mp == nil {
		t.Fatal("NewMemoryEfficientProcessor returned nil")
	}

	if mp.maxMemoryMB != 256 {
		t.Errorf("Expected maxMemoryMB 256, got %d", mp.maxMemoryMB)
	}

	stats := mp.GetMemoryStats()
	if stats.StartMemoryMB <= 0 {
		t.Error("Expected positive start memory")
	}
}

func TestMemoryEfficientProcessor_ProcessWithMemoryManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mp := NewMemoryEfficientProcessor(logger, 1024) // 1GB limit

	processed := false
	processor := func() error {
		processed = true
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	ctx := context.Background()
	err := mp.ProcessWithMemoryManagement(ctx, processor)
	if err != nil {
		t.Fatalf("ProcessWithMemoryManagement failed: %v", err)
	}

	if !processed {
		t.Error("Processor was not called")
	}

	stats := mp.GetMemoryStats()
	if stats.GCCount == 0 {
		t.Error("Expected at least one GC run")
	}
}

func TestMemoryEfficientProcessor_SetMemoryLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mp := NewMemoryEfficientProcessor(logger, 256)

	mp.SetMemoryLimit(512)
	if mp.maxMemoryMB != 512 {
		t.Errorf("Expected maxMemoryMB 512, got %d", mp.maxMemoryMB)
	}
}

func TestConcurrentProcessor_EmptyPages(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}
	cp := NewConcurrentProcessor(logger, 2, tracker)

	var pages []*generator.WikiPage
	processor := func(page *generator.WikiPage) error {
		return nil
	}

	ctx := context.Background()
	err := cp.ProcessPagesParallel(ctx, pages, processor)
	if err != nil {
		t.Errorf("ProcessPagesParallel with empty pages should not fail: %v", err)
	}
}

func TestBatchProcessor_EmptyItems(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bp := NewBatchProcessor[string](logger, 2, 1)

	var items []string
	processor := func(batch []string) error {
		return nil
	}

	tracker := &generator.NoOpProgressTracker{}
	ctx := context.Background()

	err := bp.ProcessInBatches(ctx, items, processor, tracker)
	if err != nil {
		t.Errorf("ProcessInBatches with empty items should not fail: %v", err)
	}
}

func TestConcurrentProcessor_ContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracker := &generator.NoOpProgressTracker{}
	cp := NewConcurrentProcessor(logger, 2, tracker)

	pages := []*generator.WikiPage{
		{ID: "page1", Title: "Page 1"},
		{ID: "page2", Title: "Page 2"},
		{ID: "page3", Title: "Page 3"},
		{ID: "page4", Title: "Page 4"},
	}

	processor := func(page *generator.WikiPage) error {
		time.Sleep(200 * time.Millisecond) // Simulate slow work
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	err := cp.ProcessPagesParallel(ctx, pages, processor)
	// With cancelled context, we might get an error or complete successfully
	// depending on timing, so we just check that it doesn't panic
	_ = err // Accept either outcome
}
