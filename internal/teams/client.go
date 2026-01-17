package teams

import (
	"context"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/teams"
	"github.com/pzsp-teams/teams-cli/internal/teams/creator"
)

// TeamClient provides all team-related operations
type TeamClient interface {
	// Unarchive restores an archived team
	Unarchive(ctx context.Context, teamRef string) error
	// Archive archives a team
	Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers bool) error
	// Delete removes a team
	Delete(ctx context.Context, teamRef string) error
	// Get retrieves a team
	Get(ctx context.Context, teamRef string) (*models.Team, error)
	// ListMyJoined lists teams the current user has joined
	ListMyJoined(ctx context.Context) ([]*models.Team, error)
	// Create creates teams based on the provided request data
	Create(ctx context.Context, request map[string]creator.TeamData, dryRun bool) []creator.TeamCreateResult
}

// NewTeamClient creates a new teams client
func NewTeamClient(teamsService teams.Service) TeamClient {
	return &client{
		teamsService: teamsService,
		teamCreator:  creator.NewTeamCreator(teamsService),
	}
}

type client struct {
	teamsService teams.Service
	teamCreator  creator.TeamCreator
}

// Unarchive restores an archived team
func (c *client) Unarchive(ctx context.Context, teamRef string) error {
	return c.teamsService.Unarchive(ctx, teamRef)
}

// Archive archives a team
func (c *client) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers bool) error {
	return c.teamsService.Archive(ctx, teamRef, &spoReadOnlyForMembers)
}

// Delete removes a team
func (c *client) Delete(ctx context.Context, teamRef string) error {
	return c.teamsService.Delete(ctx, teamRef)
}

// Get retrieves a team
func (c *client) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	return c.teamsService.Get(ctx, teamRef)
}

// ListMyJoined lists teams the current user has joined
func (c *client) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	return c.teamsService.ListMyJoined(ctx)
}

// Create creates teams based on the provided request data
func (c *client) Create(ctx context.Context, request map[string]creator.TeamData, dryRun bool) []creator.TeamCreateResult {
	return c.teamCreator.Create(ctx, request, dryRun)
}
