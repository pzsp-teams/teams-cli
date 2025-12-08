package client

import (
	"context"

	"github.com/pzsp-teams/lib"
)

// TeamsClient wraps the library client
type TeamsClient struct {
	*lib.Client
}

// NewTeamsClient creates a new TeamsClient by constructing the underlying library client
func NewTeamsClient(ctx context.Context, authConfig *lib.AuthConfig, senderConfig *lib.SenderConfig) (*TeamsClient, error) {
	libClient, err := lib.NewClient(ctx, authConfig, senderConfig)
	if err != nil {
		return nil, err
	}

	return &TeamsClient{
		Client: libClient,
	}, nil
}
