package common

import (
	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/client"
)

// GetTeamsClient returns the TeamsClient instance using the command's context
// This is a wrapper around client.GetOrCreateInstance for use in cmd layer
func GetTeamsClient(cmd *cobra.Command) (*client.TeamsClient, error) {
	return client.GetOrCreateInstance(cmd.Context())
}
