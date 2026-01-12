package templates

import (
	"io"
	"os"
	"testing"

	"github.com/pzsp-teams/teams-cli/internal/initializers"
	"github.com/pzsp-teams/teams-cli/internal/logger"
)

func TestMain(m *testing.M) {
	initializers.InitLogger(&logger.Config{
		Level:  logger.LevelDebug,
		Format: logger.FormatText,
		Output: io.Discard,
	})
	os.Exit(m.Run())
}
