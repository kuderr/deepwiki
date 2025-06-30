package logging

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultLogConfig(t *testing.T) {
	config := DefaultLogConfig()

	if config == nil {
		t.Fatal("DefaultLogConfig returned nil")
	}

	if config.Level != LevelInfo {
		t.Errorf("Expected default level 'info', got '%s'", config.Level)
	}

	if config.Format != "text" {
		t.Errorf("Expected default format 'text', got '%s'", config.Format)
	}

	if config.Output != "stderr" {
		t.Errorf("Expected default output 'stderr', got '%s'", config.Output)
	}

	if config.AddSource {
		t.Error("Expected AddSource to be false by default")
	}
}

func TestNewLogger_DefaultConfig(t *testing.T) {
	logger, err := NewLogger(nil)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.Logger == nil {
		t.Fatal("Logger.Logger is nil")
	}

	if logger.config == nil {
		t.Fatal("Logger.config is nil")
	}

	logger.Close()
}

func TestNewLogger_TextFormat(t *testing.T) {
	// Create a temporary file to capture output
	tempDir, err := os.MkdirTemp("", "deepwiki-text-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "text-test.log")

	config := &LogConfig{
		Level:  LevelInfo,
		Format: "text",
		Output: logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	// Write a test message
	logger.Info("test text format message")

	logger.Close()

	// Verify file was created and contains message in text format
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("Log file was not created")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logOutput := string(content)
	if !strings.Contains(logOutput, "test text format message") {
		t.Error("Log file does not contain expected message")
	}

	// Verify it's in text format (not JSON)
	if strings.Contains(logOutput, `{"time":`) {
		t.Error("Log output appears to be in JSON format, expected text format")
	}
}

func TestNewLogger_JSONFormat(t *testing.T) {
	config := &LogConfig{
		Level:  LevelDebug,
		Format: "json",
		Output: "stderr",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	logger.Close()
}

func TestNewLogger_FileOutput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "deepwiki-log-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "test.log")

	config := &LogConfig{
		Level:  LevelInfo,
		Format: "text",
		Output: logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Write a test message
	logger.Info("test message")

	logger.Close()

	// Verify file was created and contains message
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("Log file was not created")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Error("Log file does not contain expected message")
	}
}

func TestNewLogger_InvalidPath(t *testing.T) {
	config := &LogConfig{
		Level:  LevelInfo,
		Format: "text",
		Output: "/invalid/path/that/does/not/exist/test.log",
	}

	_, err := NewLogger(config)
	if err == nil {
		t.Error("Expected error for invalid log file path")
	}
}

func TestLogger_WithComponent(t *testing.T) {
	logger := GetDefaultLogger()
	componentLogger := logger.WithComponent("test-component")

	if componentLogger == nil {
		t.Fatal("WithComponent returned nil")
	}

	// Both loggers should be different instances
	if componentLogger == logger {
		t.Error("WithComponent should return a new logger instance")
	}
}

func TestLogger_WithOperation(t *testing.T) {
	logger := GetDefaultLogger()
	opLogger := logger.WithOperation("test-operation")

	if opLogger == nil {
		t.Fatal("WithOperation returned nil")
	}

	if opLogger == logger {
		t.Error("WithOperation should return a new logger instance")
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	logger := GetDefaultLogger()
	reqLogger := logger.WithRequestID("req-123")

	if reqLogger == nil {
		t.Fatal("WithRequestID returned nil")
	}

	if reqLogger == logger {
		t.Error("WithRequestID should return a new logger instance")
	}
}

func TestLogger_LogError(t *testing.T) {
	logger := GetDefaultLogger()
	ctx := context.Background()
	testError := errors.New("test error")

	// Should not panic
	logger.LogError(ctx, "test error message", testError)
}

func TestLogger_LogScanResult(t *testing.T) {
	logger := GetDefaultLogger()
	ctx := context.Background()

	// Should not panic
	logger.LogScanResult(ctx, 100, 50, time.Millisecond*200)
}

func TestLogger_LogAPICall(t *testing.T) {
	logger := GetDefaultLogger()
	ctx := context.Background()

	// Should not panic
	logger.LogAPICall(ctx, "openai", "gpt-4o", "completion", time.Second, 1000)
}

func TestLogger_LogProgress(t *testing.T) {
	logger := GetDefaultLogger()
	ctx := context.Background()

	// Should not panic
	logger.LogProgress(ctx, "generating", 5, 10)
}

func TestGlobalLogger(t *testing.T) {
	// Save original global logger
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// Reset global logger for this test
	globalLogger = nil

	// Test default global logger
	logger1 := GetGlobalLogger()
	if logger1 == nil {
		t.Fatal("GetGlobalLogger returned nil")
	}

	// Should return same instance
	logger2 := GetGlobalLogger()
	if logger1 != logger2 {
		t.Error("GetGlobalLogger should return same instance")
	}

	// Test setting custom global logger
	customLogger := GetDefaultLogger()
	SetGlobalLogger(customLogger)

	logger3 := GetGlobalLogger()
	if logger3 != customLogger {
		t.Error("GetGlobalLogger should return custom logger after SetGlobalLogger")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Should not panic
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	ctx := context.Background()
	DebugContext(ctx, "debug context message")
	InfoContext(ctx, "info context message")
	WarnContext(ctx, "warn context message")
	ErrorContext(ctx, "error context message")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelDebug, "debug"},
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
	}

	for _, test := range tests {
		if string(test.level) != test.expected {
			t.Errorf("Expected level '%s', got '%s'", test.expected, string(test.level))
		}
	}
}

func TestLogger_Close(t *testing.T) {
	// Test close with stdout/stderr (should not error for standard outputs)
	config := &LogConfig{
		Level:  LevelInfo,
		Format: "text",
		Output: "stdout",
	}

	logger1, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Close should not return an error for stdout/stderr
	err = logger1.Close()
	if err != nil {
		t.Errorf("Close failed for stdout logger: %v", err)
	}

	// Test close with file logger
	tempDir, err := os.MkdirTemp("", "deepwiki-close-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "close-test.log")
	config2 := &LogConfig{
		Level:  LevelInfo,
		Format: "text",
		Output: logFile,
	}

	logger2, err := NewLogger(config2)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	err = logger2.Close()
	if err != nil {
		t.Errorf("Close failed for file logger: %v", err)
	}
}

func TestLogConfig_TimeFormat(t *testing.T) {
	config := &LogConfig{
		Level:      LevelInfo,
		Format:     "text",
		Output:     "stderr",
		TimeFormat: time.RFC822,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	if logger.config.TimeFormat != time.RFC822 {
		t.Errorf("Expected time format %s, got %s", time.RFC822, logger.config.TimeFormat)
	}

	logger.Close()
}

func TestLogConfig_AddSource(t *testing.T) {
	config := &LogConfig{
		Level:     LevelInfo,
		Format:    "text",
		Output:    "stderr",
		AddSource: true,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	if !logger.config.AddSource {
		t.Error("Expected AddSource to be true")
	}

	logger.Close()
}
