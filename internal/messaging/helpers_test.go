package messaging

import (
	"os"
	"testing"

	"github.com/pzsp-teams/cli/internal/initializers"
)

func TestMain(m *testing.M) {
	initializers.SetupTestLogger()
	os.Exit(m.Run())
}
