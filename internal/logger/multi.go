package logger

// MultiLogger dispatches log calls to multiple underlying loggers.
type MultiLogger struct {
	loggers []Logger
}

// NewMultiLogger creates a new MultiLogger that dispatches to all provided loggers.
// Each logger can have its own configuration (level, format, output).
// Log messages are sent to all loggers, and each logger applies its own filtering.
func NewMultiLogger(loggers ...Logger) Logger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// Debug logs a debug-level message to all underlying loggers.
func (m *MultiLogger) Debug(msg string, args ...any) {
	for _, l := range m.loggers {
		l.Debug(msg, args...)
	}
}

// Info logs an info-level message to all underlying loggers.
func (m *MultiLogger) Info(msg string, args ...any) {
	for _, l := range m.loggers {
		l.Info(msg, args...)
	}
}

// Warn logs a warning-level message to all underlying loggers.
func (m *MultiLogger) Warn(msg string, args ...any) {
	for _, l := range m.loggers {
		l.Warn(msg, args...)
	}
}

// Error logs an error-level message to all underlying loggers.
func (m *MultiLogger) Error(msg string, args ...any) {
	for _, l := range m.loggers {
		l.Error(msg, args...)
	}
}

// With returns a new MultiLogger with the given key-value pairs added to the context
// of all underlying loggers.
func (m *MultiLogger) With(args ...any) Logger {
	newLoggers := make([]Logger, len(m.loggers))
	for i, l := range m.loggers {
		newLoggers[i] = l.With(args...)
	}
	return &MultiLogger{
		loggers: newLoggers,
	}
}

// GetLoggers returns a copy of the current loggers (for identification).
func (m *MultiLogger) GetLoggers() []Logger {
	result := make([]Logger, len(m.loggers))
	copy(result, m.loggers)
	return result
}

// RemoveLoggerAt removes the logger at the specified index.
func (m *MultiLogger) RemoveLoggerAt(index int) {
	if index < 0 || index >= len(m.loggers) {
		return
	}
	m.loggers = append(m.loggers[:index], m.loggers[index+1:]...)
}
