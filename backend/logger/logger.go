package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger is a structured logger handler
type Logger struct {
	logger *slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  LogLevel
	Format string // "json" or "text"
	Output io.Writer
}

// New creates a new logger with the given configuration
func New(config Config) *Logger {
	// Set default output if not provided
	if config.Output == nil {
		config.Output = os.Stdout
	}

	// Parse log level
	var level slog.Level
	switch strings.ToLower(string(config.Level)) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler based on format
	var handler slog.Handler
	if strings.ToLower(config.Format) == "json" {
		handler = slog.NewJSONHandler(config.Output, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(config.Output, &slog.HandlerOptions{
			Level: level,
		})
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Default creates a logger with default configuration
func Default() *Logger {
	return New(Config{
		Level:  LevelInfo,
		Format: "json",
		Output: os.Stdout,
	})
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With adds attributes to the logger
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithGroup adds a group to the logger
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		logger: l.logger.WithGroup(name),
	}
}

// RequestID extracts request ID from context and adds it to the log
func (l *Logger) RequestID(ctx context.Context) *Logger {
	if requestID := ctx.Value("request_id"); requestID != nil {
		return l.With("request_id", requestID)
	}
	return l
}

// StoreID extracts store ID from context and adds it to the log
func (l *Logger) StoreID(ctx context.Context) *Logger {
	if storeID := ctx.Value("store_id"); storeID != nil {
		return l.With("store_id", storeID)
	}
	return l
}

// UserID extracts user ID from context and adds it to the log
func (l *Logger) UserID(ctx context.Context) *Logger {
	if userID := ctx.Value("user_id"); userID != nil {
		return l.With("user_id", userID)
	}
	return l
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path string, statusCode int, duration int64, ctx context.Context) {
	l.RequestID(ctx).StoreID(ctx).UserID(ctx).Info(
		"HTTP request",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration,
	)
}

// LogError logs an error with context
func (l *Logger) LogError(err error, msg string, ctx context.Context) {
	l.RequestID(ctx).StoreID(ctx).UserID(ctx).Error(
		msg,
		"error", err.Error(),
		"error_type", fmt.Sprintf("%T", err),
	)
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config Config) {
	globalLogger = New(config)
}

// Debug logs a debug message using the global logger
func Debug(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Debug(msg, args...)
	}
}

// Info logs an info message using the global logger
func Info(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Info(msg, args...)
	}
}

// Warn logs a warning message using the global logger
func Warn(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Warn(msg, args...)
	}
}

// Error logs an error message using the global logger
func Error(msg string, args ...any) {
	if globalLogger != nil {
		globalLogger.Error(msg, args...)
	}
}

// With adds attributes to the global logger
func With(args ...any) *Logger {
	if globalLogger != nil {
		return globalLogger.With(args...)
	}
	return Default().With(args...)
}
