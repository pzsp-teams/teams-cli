package common

import (
	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// GetTeamsClient returns the TeamsClient instance using the command's context
// This is a wrapper around client.GetOrCreateInstance for use in cmd layer
func GetTeamsClient(cmd *cobra.Command) (*client.TeamsClient, error) {
	log := initializers.Logger
	log.Debug("Creating Teams client")

	teamsClient, err := client.GetOrCreateInstance(cmd.Context())
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
	}

	return teamsClient, err
}
