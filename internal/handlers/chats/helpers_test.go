package chats

import (
	"testing"

	"github.com/pzsp-teams/teams-cli/internal/initializers"
)

func TestMain(m *testing.M) {
	initializers.SetupTestLogger()
	m.Run()
}
