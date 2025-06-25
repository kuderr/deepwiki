package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogConfig represents logging configuration
type LogConfig struct {
	Level      LogLevel `yaml:"level" json:"level"`
	Format     string   `yaml:"format" json:"format"` // "text" or "json"
	Output     string   `yaml:"output" json:"output"` // "stdout", "stderr", or file path
	AddSource  bool     `yaml:"add_source" json:"addSource"`
	TimeFormat string   `yaml:"time_format" json:"timeFormat"`
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      LevelInfo,
		Format:     "text",
		Output:     "stderr",
		AddSource:  false,
		TimeFormat: time.RFC3339,
	}
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	config *LogConfig
	output io.Writer
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config *LogConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// Determine output writer
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr", "":
		output = os.Stderr
	default:
		// File output
		dir := filepath.Dir(config.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		output = file
	}

	// Configure slog level
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(config.TimeFormat))
				}
			}
			return a
		},
	}

	// Create handler based on format
	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		config: config,
		output: output,
	}, nil
}

// GetDefaultLogger returns a default logger for the application
func GetDefaultLogger() *Logger {
	logger, _ := NewLogger(DefaultLogConfig())
	return logger
}

// WithComponent returns a logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		config: l.config,
		output: l.output,
	}
}

// WithOperation returns a logger with an operation field
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
		config: l.config,
		output: l.output,
	}
}

// WithRequestID returns a logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
		config: l.config,
		output: l.output,
	}
}

// LogError logs an error with additional context
func (l *Logger) LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	l.Logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// LogScanResult logs scan result statistics
func (l *Logger) LogScanResult(ctx context.Context, totalFiles, filteredFiles int, scanTime time.Duration) {
	l.Logger.LogAttrs(ctx, slog.LevelInfo, "directory scan completed",
		slog.Int("total_files", totalFiles),
		slog.Int("filtered_files", filteredFiles),
		slog.Duration("scan_time", scanTime),
	)
}

// LogAPICall logs an API call with timing
func (l *Logger) LogAPICall(ctx context.Context, provider, model, operation string, duration time.Duration, tokens int) {
	l.Logger.LogAttrs(ctx, slog.LevelInfo, "api call completed",
		slog.String("provider", provider),
		slog.String("model", model),
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.Int("tokens", tokens),
	)
}

// LogProgress logs progress updates
func (l *Logger) LogProgress(ctx context.Context, operation string, current, total int) {
	l.Logger.LogAttrs(ctx, slog.LevelInfo, "progress update",
		slog.String("operation", operation),
		slog.Int("current", current),
		slog.Int("total", total),
		slog.Float64("percent", float64(current)/float64(total)*100),
	)
}

// Close closes the logger (if it's writing to a file)
func (l *Logger) Close() error {
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Global logger instance
var globalLogger *Logger

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = GetDefaultLogger()
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, args ...any) {
	GetGlobalLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	GetGlobalLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	GetGlobalLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	GetGlobalLogger().Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	GetGlobalLogger().ErrorContext(ctx, msg, args...)
}
