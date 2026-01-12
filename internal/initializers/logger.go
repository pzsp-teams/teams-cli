package initializers

import (
	"io"
	"os"

	"github.com/pzsp-teams/cli/internal/logger"
)

// Logger is the global logger instance that can be used throughout the application.
var Logger logger.Logger

// StderrLogger tracks the stderr logger for potential removal in TUI mode.
var StderrLogger logger.Logger

// InitLogger initializes the global logger with the provided configurations.
// Each config creates a separate logger, and all are combined into a MultiLogger.
func InitLogger(configs ...*logger.Config) {
	if len(configs) == 0 {
		InitDefaultLogger()
		return
	}

	if len(configs) == 1 {
		Logger = logger.NewCharmFromConfig(configs[0])
		return
	}

	loggers := make([]logger.Logger, len(configs))
	for i, cfg := range configs {
		loggers[i] = logger.NewCharmFromConfig(cfg)
	}
	Logger = logger.NewMultiLogger(loggers...)
}

// InitDefaultLogger initializes basic logger
func InitDefaultLogger() {
	Logger = logger.NewCharmFromConfig(logger.DefaultConfig())
}

// MultiOutputConfig holds configuration for initializing a multi-output logger.
type MultiOutputConfig struct {
	StderrLevel         logger.Level
	FileLevel           logger.Level
	FileWriter          io.Writer
	StderrOmitTimestamp bool
	FileOmitTimestamp   bool
	StderrAddSource     bool
	FileAddSource       bool
}

// InitMultiOutputLogger creates a MultiLogger with one logger for stderr and one for a file.
func InitMultiOutputLogger(cfg MultiOutputConfig) {
	stderrLogger := logger.NewCharmFromConfig(&logger.Config{
		Level:         cfg.StderrLevel,
		Format:        logger.FormatText,
		Output:        os.Stderr,
		OmitTimestamp: cfg.StderrOmitTimestamp,
		AddSource:     cfg.StderrAddSource,
	})

	fileLogger := logger.NewCharmFromConfig(&logger.Config{
		Level:         cfg.FileLevel,
		Format:        logger.FormatText,
		Output:        cfg.FileWriter,
		OmitTimestamp: cfg.FileOmitTimestamp,
		AddSource:     cfg.FileAddSource,
	})

	Logger = logger.NewMultiLogger(stderrLogger, fileLogger)
}

// DisableStderrLogger removes the stderr logger from the MultiLogger.
// This is useful for TUI mode to prevent log output from interfering with the display.
func DisableStderrLogger() {
	if multiLogger, ok := Logger.(*logger.MultiLogger); ok {
		loggers := multiLogger.GetLoggers()
		for i, l := range loggers {
			if l == StderrLogger {
				multiLogger.RemoveLoggerAt(i)
				return
			}
		}
	}
}
