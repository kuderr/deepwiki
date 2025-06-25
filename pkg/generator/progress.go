package generator

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ConsoleProgressTracker implements ProgressTracker with console output
type ConsoleProgressTracker struct {
	logger      *slog.Logger
	currentTask string
	total       int
	current     int
	startTime   time.Time
	mu          sync.RWMutex
}

// NewConsoleProgressTracker creates a new console progress tracker
func NewConsoleProgressTracker(logger *slog.Logger) *ConsoleProgressTracker {
	return &ConsoleProgressTracker{
		logger: logger.With("component", "progress"),
	}
}

// StartTask starts a new task with progress tracking
func (p *ConsoleProgressTracker) StartTask(taskName string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentTask = taskName
	p.total = total
	p.current = 0
	p.startTime = time.Now()

	p.logger.Info("Started task",
		"task", taskName,
		"total", total)
}

// UpdateProgress updates the current progress
func (p *ConsoleProgressTracker) UpdateProgress(current int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current

	percentage := float64(current) / float64(p.total) * 100
	elapsed := time.Since(p.startTime)

	// Estimate remaining time
	var eta time.Duration
	if current > 0 {
		avgTimePerItem := elapsed / time.Duration(current)
		remainingItems := p.total - current
		eta = avgTimePerItem * time.Duration(remainingItems)
	}

	p.logger.Info("Progress update",
		"task", p.currentTask,
		"current", current,
		"total", p.total,
		"percentage", fmt.Sprintf("%.1f%%", percentage),
		"elapsed", elapsed.Round(time.Second),
		"eta", eta.Round(time.Second),
		"message", message)
}

// CompleteTask marks the current task as completed
func (p *ConsoleProgressTracker) CompleteTask(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	duration := time.Since(p.startTime)

	p.logger.Info("Task completed",
		"task", p.currentTask,
		"total", p.total,
		"duration", duration.Round(time.Second),
		"message", message)

	// Reset state
	p.currentTask = ""
	p.total = 0
	p.current = 0
}

// SetError logs an error during task execution
func (p *ConsoleProgressTracker) SetError(err error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.logger.Error("Task error",
		"task", p.currentTask,
		"current", p.current,
		"total", p.total,
		"error", err)
}

// GetCurrentProgress returns the current progress information
func (p *ConsoleProgressTracker) GetCurrentProgress() (string, int, int, time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var elapsed time.Duration
	if !p.startTime.IsZero() {
		elapsed = time.Since(p.startTime)
	}

	return p.currentTask, p.current, p.total, elapsed
}

// BarProgressTracker implements ProgressTracker with a progress bar
type BarProgressTracker struct {
	logger      *slog.Logger
	currentTask string
	total       int
	current     int
	startTime   time.Time
	barWidth    int
	mu          sync.RWMutex
}

// NewBarProgressTracker creates a new bar progress tracker
func NewBarProgressTracker(logger *slog.Logger) *BarProgressTracker {
	return &BarProgressTracker{
		logger:   logger.With("component", "progress"),
		barWidth: 50,
	}
}

// StartTask starts a new task with progress tracking
func (p *BarProgressTracker) StartTask(taskName string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentTask = taskName
	p.total = total
	p.current = 0
	p.startTime = time.Now()

	fmt.Printf("\n🚀 %s (0/%d)\n", taskName, total)
	p.drawProgressBar()
}

// UpdateProgress updates the current progress
func (p *BarProgressTracker) UpdateProgress(current int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.drawProgressBar()

	if message != "" {
		fmt.Printf("\r%s", message)
	}
}

// CompleteTask marks the current task as completed
func (p *BarProgressTracker) CompleteTask(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	p.drawProgressBar()

	duration := time.Since(p.startTime)
	fmt.Printf("\n✅ %s completed in %v\n", p.currentTask, duration.Round(time.Second))

	if message != "" {
		fmt.Printf("   %s\n", message)
	}

	// Reset state
	p.currentTask = ""
	p.total = 0
	p.current = 0
}

// SetError logs an error during task execution
func (p *BarProgressTracker) SetError(err error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Printf("\n❌ Error in %s: %v\n", p.currentTask, err)
}

// drawProgressBar draws the progress bar
func (p *BarProgressTracker) drawProgressBar() {
	if p.total <= 0 {
		return
	}

	percentage := float64(p.current) / float64(p.total)
	filled := int(percentage * float64(p.barWidth))

	bar := "["
	for i := 0; i < p.barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled && p.current < p.total {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	elapsed := time.Since(p.startTime)
	var eta string
	if p.current > 0 && p.current < p.total {
		avgTimePerItem := elapsed / time.Duration(p.current)
		remainingItems := p.total - p.current
		etaDuration := avgTimePerItem * time.Duration(remainingItems)
		eta = fmt.Sprintf(" ETA: %v", etaDuration.Round(time.Second))
	}

	fmt.Printf("\r%s %3.0f%% (%d/%d) %v%s",
		bar,
		percentage*100,
		p.current,
		p.total,
		elapsed.Round(time.Second),
		eta)
}
