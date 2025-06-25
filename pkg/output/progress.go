package output

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// EnhancedProgressTracker provides advanced progress tracking with visual bars
type EnhancedProgressTracker struct {
	logger      *slog.Logger
	currentTask string
	total       int
	current     int
	startTime   time.Time
	lastUpdate  time.Time
	barWidth    int
	showETA     bool
	showRate    bool
	mu          sync.RWMutex
	isTerminal  bool
	termWidth   int
}

// NewEnhancedProgressTracker creates a new enhanced progress tracker
func NewEnhancedProgressTracker(logger *slog.Logger) *EnhancedProgressTracker {
	tracker := &EnhancedProgressTracker{
		logger:     logger.With("component", "progress"),
		barWidth:   40,
		showETA:    true,
		showRate:   true,
		isTerminal: isTerminal(),
		termWidth:  getTerminalWidth(),
	}

	// Adjust bar width based on terminal size
	if tracker.isTerminal && tracker.termWidth > 0 {
		// Reserve space for text: "[100%] (1000/1000) 1m23s ETA: 0s 123.4/s "
		availableWidth := tracker.termWidth - 50
		if availableWidth > 20 && availableWidth < 80 {
			tracker.barWidth = availableWidth
		}
	}

	return tracker
}

// StartTask starts tracking a new task
func (p *EnhancedProgressTracker) StartTask(taskName string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentTask = taskName
	p.total = total
	p.current = 0
	p.startTime = time.Now()
	p.lastUpdate = time.Now()

	if p.isTerminal {
		fmt.Printf("\n🚀 %s\n", taskName)
		p.drawEnhancedBar()
	} else {
		p.logger.Info("Started task", "task", taskName, "total", total)
	}
}

// UpdateProgress updates the progress
func (p *EnhancedProgressTracker) UpdateProgress(current int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	now := time.Now()

	// Update display at most once per 100ms to avoid flickering
	if now.Sub(p.lastUpdate) < 100*time.Millisecond && current < p.total {
		return
	}
	p.lastUpdate = now

	if p.isTerminal {
		p.drawEnhancedBar()
		if message != "" && len(message) < 50 {
			fmt.Printf(" %s", message)
		}
	} else {
		percentage := float64(current) / float64(p.total) * 100
		p.logger.Info("Progress",
			"task", p.currentTask,
			"current", current,
			"total", p.total,
			"percentage", fmt.Sprintf("%.1f%%", percentage),
			"message", message)
	}
}

// CompleteTask marks the task as completed
func (p *EnhancedProgressTracker) CompleteTask(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	duration := time.Since(p.startTime)

	if p.isTerminal {
		// Clear the line and show completion
		fmt.Print("\r" + strings.Repeat(" ", p.termWidth-1) + "\r")
		fmt.Printf("✅ %s completed in %v", p.currentTask, duration.Round(time.Second))
		if message != "" {
			fmt.Printf(" - %s", message)
		}
		fmt.Println()
	} else {
		p.logger.Info("Task completed",
			"task", p.currentTask,
			"duration", duration.Round(time.Second),
			"message", message)
	}

	// Reset state
	p.currentTask = ""
	p.total = 0
	p.current = 0
}

// SetError reports an error
func (p *EnhancedProgressTracker) SetError(err error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isTerminal {
		fmt.Printf("\n❌ Error in %s: %v\n", p.currentTask, err)
	} else {
		p.logger.Error("Task error",
			"task", p.currentTask,
			"error", err)
	}
}

// drawEnhancedBar draws an enhanced progress bar
func (p *EnhancedProgressTracker) drawEnhancedBar() {
	if p.total <= 0 {
		return
	}

	percentage := float64(p.current) / float64(p.total)
	filled := int(percentage * float64(p.barWidth))

	// Build the bar
	var bar strings.Builder
	bar.WriteString("[")

	for i := 0; i < p.barWidth; i++ {
		switch {
		case i < filled:
			bar.WriteString("█")
		case i == filled && p.current < p.total:
			bar.WriteString("▓")
		default:
			bar.WriteString("░")
		}
	}
	bar.WriteString("]")

	// Calculate timing information
	elapsed := time.Since(p.startTime)
	var eta string
	var rate string

	if p.current > 0 {
		if p.showRate {
			itemsPerSecond := float64(p.current) / elapsed.Seconds()
			rate = fmt.Sprintf(" %.1f/s", itemsPerSecond)
		}

		if p.showETA && p.current < p.total {
			avgTimePerItem := elapsed / time.Duration(p.current)
			remainingItems := p.total - p.current
			etaDuration := avgTimePerItem * time.Duration(remainingItems)
			eta = fmt.Sprintf(" ETA: %v", etaDuration.Round(time.Second))
		}
	}

	// Format the complete line
	line := fmt.Sprintf("\r%s %3.0f%% (%d/%d) %v%s%s",
		bar.String(),
		percentage*100,
		p.current,
		p.total,
		elapsed.Round(time.Second),
		eta,
		rate)

	fmt.Print(line)
}

// MultiTaskTracker tracks multiple concurrent tasks
type MultiTaskTracker struct {
	logger    *slog.Logger
	tasks     map[string]*TaskInfo
	mu        sync.RWMutex
	startTime time.Time
}

// TaskInfo holds information about a single task
type TaskInfo struct {
	Name      string
	Current   int
	Total     int
	StartTime time.Time
	Status    TaskStatus
}

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
)

// NewMultiTaskTracker creates a new multi-task tracker
func NewMultiTaskTracker(logger *slog.Logger) *MultiTaskTracker {
	return &MultiTaskTracker{
		logger:    logger.With("component", "multi-progress"),
		tasks:     make(map[string]*TaskInfo),
		startTime: time.Now(),
	}
}

// AddTask adds a new task to track
func (m *MultiTaskTracker) AddTask(taskID, taskName string, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks[taskID] = &TaskInfo{
		Name:      taskName,
		Current:   0,
		Total:     total,
		StartTime: time.Now(),
		Status:    TaskPending,
	}
}

// StartTask starts tracking a task
func (m *MultiTaskTracker) StartTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, exists := m.tasks[taskID]; exists {
		task.Status = TaskRunning
		task.StartTime = time.Now()
	}
}

// UpdateTask updates task progress
func (m *MultiTaskTracker) UpdateTask(taskID string, current int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, exists := m.tasks[taskID]; exists {
		task.Current = current
		if current >= task.Total {
			task.Status = TaskCompleted
		}
	}
}

// CompleteTask marks a task as completed
func (m *MultiTaskTracker) CompleteTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, exists := m.tasks[taskID]; exists {
		task.Current = task.Total
		task.Status = TaskCompleted
	}
}

// FailTask marks a task as failed
func (m *MultiTaskTracker) FailTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, exists := m.tasks[taskID]; exists {
		task.Status = TaskFailed
	}
}

// PrintSummary prints a summary of all tasks
func (m *MultiTaskTracker) PrintSummary() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Println("\n📊 Task Summary:")
	for _, task := range m.tasks {
		status := "❓"
		switch task.Status {
		case TaskPending:
			status = "⏳"
		case TaskRunning:
			status = "🔄"
		case TaskCompleted:
			status = "✅"
		case TaskFailed:
			status = "❌"
		}

		fmt.Printf("  %s %s: %d/%d\n", status, task.Name, task.Current, task.Total)
	}

	totalDuration := time.Since(m.startTime)
	fmt.Printf("  ⏱️  Total time: %v\n", totalDuration.Round(time.Second))
}

// Utility functions for terminal detection

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	// Check if stdout is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// getTerminalWidth gets the width of the terminal
func getTerminalWidth() int {
	// Default width if we can't determine
	defaultWidth := 80

	// Try to get terminal size (Unix-like systems)
	if width := getTerminalWidthUnix(); width > 0 {
		return width
	}

	return defaultWidth
}

// getTerminalWidthUnix gets terminal width on Unix-like systems
func getTerminalWidthUnix() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		_ = errno // Ignore error
		return 0
	}

	return int(ws.Col)
}
