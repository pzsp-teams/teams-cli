package channelcreation

import (
	"testing"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/logger"
)

type noopLogger struct{}

func (noopLogger) With(args ...any) logger.Logger {
	return noopLogger{}
}

func (noopLogger) Debug(msg string, args ...any) {}
func (noopLogger) Info(msg string, args ...any)  {}
func (noopLogger) Warn(msg string, args ...any)  {}
func (noopLogger) Error(msg string, args ...any) {}

var _ logger.Logger = noopLogger{}

func TestMain(m *testing.M) {
	initializers.Logger = noopLogger{}
	m.Run()
}
