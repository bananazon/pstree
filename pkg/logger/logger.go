package logger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

var (
	Logger *slog.Logger
	once   sync.Once
)

type CustomHandler struct {
	level slog.Level
}

// Enabled determines if a log record at the given level should be processed.
//
// This method implements the slog.Handler interface and is called to check if a log
// record at the specified level should be handled. It returns true if the record's
// level is greater than or equal to the handler's configured level.
//
// Parameters:
//   - _: Context (unused)
//   - level: The log level to check
//
// Returns:
//   - bool: true if the record should be processed, false otherwise
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle processes a log record by formatting and printing it.
//
// This method implements the slog.Handler interface and is called to process a log record.
// It formats the record with its level and message and prints it to standard output.
//
// Parameters:
//   - _: Context (unused)
//   - r: The log record to process
//
// Returns:
//   - error: nil if successful, or an error if the record could not be processed
func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	fmt.Printf("[%s] %s\n", r.Level, r.Message)
	return nil
}

// WithAttrs returns a new handler with the given attributes.
//
// This method implements the slog.Handler interface. In this simple implementation,
// it ignores the attributes and returns the same handler.
//
// Parameters:
//   - attrs: Attributes to add to the handler (ignored in this implementation)
//
// Returns:
//   - slog.Handler: The same handler (attributes are ignored)
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns a new handler with the given group name.
//
// This method implements the slog.Handler interface. In this simple implementation,
// it ignores the group name and returns the same handler.
//
// Parameters:
//   - name: Group name to add to the handler (ignored in this implementation)
//
// Returns:
//   - slog.Handler: The same handler (group name is ignored)
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}

// Init initializes the global logger with the specified log level.
//
// This function creates a new logger with a CustomHandler configured at the specified level.
// It uses sync.Once to ensure that the logger is only initialized once, making it safe for
// concurrent use.
//
// Parameters:
//   - level: The minimum log level to process (e.g., slog.LevelDebug, slog.LevelInfo)
func Init(level slog.Level) {
	once.Do(func() {
		Logger = slog.New(&CustomHandler{level: level})
	})
}
