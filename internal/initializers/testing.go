//go:build !release

package initializers

import "github.com/pzsp-teams/cli/internal/logger"

type noopLogger struct{}

// With returns a new logger with additional context fields. For noopLogger, this is a no-op.
func (noopLogger) With(args ...any) logger.Logger { return noopLogger{} }

// Debug logs a debug message. For noopLogger, this is a no-op.
func (noopLogger) Debug(msg string, args ...any) {}

// Info logs an info message. For noopLogger, this is a no-op.
func (noopLogger) Info(msg string, args ...any) {}

// Warn logs a warning message. For noopLogger, this is a no-op.
func (noopLogger) Warn(msg string, args ...any) {}

// Error logs an error message. For noopLogger, this is a no-op.
func (noopLogger) Error(msg string, args ...any) {}

var _ logger.Logger = noopLogger{}

// SetupTestLogger configures the global Logger to use a no-op implementation for testing.
func SetupTestLogger() {
	Logger = noopLogger{}
}
