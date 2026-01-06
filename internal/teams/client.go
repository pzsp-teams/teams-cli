package teams

import (
	"context"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/teams"
)

// NewClient creates a new teams client
func NewClient(teamsService teams.Service) *Client {
	return &Client{
		teamsService,
	}
}

// Client provides all team-related operations
type Client struct {
	teamsService teams.Service
}

// Unarchive restores an archived team
func (c *Client) Unarchive(ctx context.Context, teamRef string) error {
	return c.teamsService.Unarchive(ctx, teamRef)
}

// Archive archives a team
func (c *Client) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers bool) error {
	return c.teamsService.Archive(ctx, teamRef, &spoReadOnlyForMembers)
}

// Delete removes a team
func (c *Client) Delete(ctx context.Context, teamRef string) error {
	return c.teamsService.Delete(ctx, teamRef)
}

// CreateFromTemplate creates a team from a template
// func (c *Client) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
// 	return c.teamsService.CreateFromTemplate(ctx, displayName, description, owners)
// }

// Get retrieves a team
func (c *Client) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	return c.teamsService.Get(ctx, teamRef)
}

// ListMyJoined lists teams the current user has joined
func (c *Client) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	return c.teamsService.ListMyJoined(ctx)
}
