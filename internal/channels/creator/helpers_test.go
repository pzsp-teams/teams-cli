package creator

import (
	"testing"

	"github.com/pzsp-teams/cli/internal/initializers"
)

func TestMain(m *testing.M) {
	initializers.SetupTestLogger()
	m.Run()
}
