package creator

import (
	"context"
	"fmt"

	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/teams"
)

type teamCreator struct {
	ts teams.Service
}

// TeamCreator defines the interface for batch team creation operations.
type TeamCreator interface {
	Create(ctx context.Context, request map[string]TeamData, dryRun bool) []TeamCreateResult
}

// NewTeamCreator creates a new TeamCreator instance with the provided teams service.
func NewTeamCreator(ts teams.Service) TeamCreator {
	return &teamCreator{
		ts: ts,
	}
}

// Create creates teams based on given request data
func (tc *teamCreator) Create(ctx context.Context, request map[string]TeamData, dryRun bool) []TeamCreateResult {
	logger := initializers.Logger
	logger.Info("Starting teams creation")

	bodies := tc.transformRequestToCreateTeamBody(request)
	actions := tc.planActions(ctx, bodies)

	var results []TeamCreateResult
	if dryRun {
		logger.Info("Dry run enabled - no teams will be created")
		results = tc.dryRunActions(actions)
	} else {
		results = tc.executeActions(ctx, actions)
	}

	logger.Info("Teams creation process completed")
	return results
}

func (tc *teamCreator) executeActions(ctx context.Context, actions []action) []TeamCreateResult {
	return corecreator.ExecuteActions(
		ctx,
		actions,
		tc.nilResultFactory,
		logExecutionResult,
	)
}

func (tc *teamCreator) nilResultFactory(act corecreator.Action[createTeamBody, TeamCreateResult]) TeamCreateResult {
	logger := initializers.Logger
	logger.Error("Action returned nil result", "team", act.Body.DisplayName)
	return TeamCreateResult{
		TeamName: act.Body.DisplayName,
		TeamID:   "",
		Error:    fmt.Errorf("action returned nil result"),
		Status:   corecreator.StatusFailed,
	}
}

func (tc *teamCreator) dryRunActions(actions []action) []TeamCreateResult {
	return corecreator.DryRunActions(
		actions,
		tc.nilResultFactory,
		logDryRunResult,
	)
}

func (tc *teamCreator) planActions(ctx context.Context, bodies []createTeamBody) []action {
	return corecreator.PlanActions(
		ctx,
		bodies,
		tc.createTeamAction,
	)
}

func (tc *teamCreator) createTeamAction(ctx context.Context, body *createTeamBody) action {
	return action{
		Body: *body,
		Run: func(ctx context.Context, body createTeamBody) *TeamCreateResult {
			teamID, err := tc.ts.CreateFromTemplate(
				ctx,
				body.DisplayName,
				body.Description,
				body.OwnerRefs,
				body.MemberRefs,
				body.Visibility,
				body.IncludeMe,
			)
			if err != nil {
				err = fmt.Errorf("failed to create team %s: %w", body.DisplayName, err)
				return &TeamCreateResult{
					TeamName:    body.DisplayName,
					TeamID:      "",
					Error:       err,
					Status:      corecreator.StatusFailed,
					Description: body.Description,
					Visibility:  body.Visibility,
				}
			}
			return &TeamCreateResult{
				TeamName:    body.DisplayName,
				TeamID:      teamID,
				Error:       nil,
				Status:      corecreator.StatusCreated,
				MemberRefs:  body.MemberRefs,
				OwnerRefs:   body.OwnerRefs,
				Description: body.Description,
				Visibility:  body.Visibility,
			}
		},
		Result: &TeamCreateResult{
			TeamName:    body.DisplayName,
			TeamID:      "",
			Error:       nil,
			Status:      corecreator.StatusWouldCreate,
			MemberRefs:  body.MemberRefs,
			OwnerRefs:   body.OwnerRefs,
			Description: body.Description,
			Visibility:  body.Visibility,
		},
	}
}

func (tc *teamCreator) transformRequestToCreateTeamBody(data map[string]TeamData) []createTeamBody {
	out := make([]createTeamBody, 0, len(data))

	for displayName, teamData := range data {
		visibility := teamData.Visibility
		if visibility == "" {
			visibility = "private"
		}
		body := createTeamBody{
			DisplayName: displayName,
			Description: teamData.Description,
			MemberRefs:  teamData.Members,
			OwnerRefs:   teamData.Owners,
			Visibility:  visibility,
			IncludeMe:   teamData.IncludeMe,
		}
		out = append(out, body)
	}

	return out
}

func logExecutionResult(result *TeamCreateResult) {
	logger := initializers.Logger
	switch result.Status {
	case corecreator.StatusCreated:
		logger.Info("Team created successfully", "team", result.TeamName, "team_id", result.TeamID, "status", result.Status, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case corecreator.StatusFailed:
		logger.Error("Team operation failed", "team", result.TeamName, "error", result.Error, "status", result.Status)
	}
}

func logDryRunResult(result *TeamCreateResult) {
	logger := initializers.Logger
	switch result.Status {
	case corecreator.StatusWouldCreate:
		logger.Info("Dry run: Team would be created", "team", result.TeamName, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case corecreator.StatusFailed:
		logger.Error("Dry run: Team creation would fail", "team", result.TeamName, "error", result.Error)
	}
}
