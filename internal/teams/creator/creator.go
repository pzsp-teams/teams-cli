package creator

import (
	"context"
	"fmt"
	"strings"

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
		tc.planActionForBody,
	)
}

func (tc *teamCreator) planActionForBody(ctx context.Context, body *createTeamBody) action {
	exists, err := tc.checkTeamExists(ctx, body.DisplayName)
	if err != nil {
		errToShow := fmt.Errorf("failed to check existence of team %s: %w", body.DisplayName, err)
		return staticAction(body, failedResult(errToShow, body))
	}

	if exists {
		return staticAction(body, alreadyExistsResult(body))
	}

	return tc.createTeamAction(body)
}

func (tc *teamCreator) createTeamAction(body *createTeamBody) action {
	return action{
		Body: *body,
		Run: func(ctx context.Context, body createTeamBody) *TeamCreateResult {
			teamID, err := tc.ts.CreateFromTemplate(
				ctx,
				body.DisplayName,
				body.Description,
				body.OwnerRefs,
				body.MemberRefs,
				"private",
				false,
			)
			if err != nil {
				err = fmt.Errorf("failed to create team %s: %w", body.DisplayName, err)
				return &TeamCreateResult{
					TeamName:    body.DisplayName,
					TeamID:      "",
					Error:       err,
					Status:      corecreator.StatusFailed,
					Description: body.Description,
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
		},
	}
}

func (tc *teamCreator) transformRequestToCreateTeamBody(data map[string]TeamData) []createTeamBody {
	out := make([]createTeamBody, 0, len(data))

	for displayName, teamData := range data {
		body := createTeamBody{
			DisplayName: displayName,
			Description: extractDescription(teamData),
			MemberRefs:  extractStringSlice(teamData, "members"),
			OwnerRefs:   extractStringSlice(teamData, "owners"),
		}
		out = append(out, body)
	}

	return out
}

func (tc *teamCreator) checkTeamExists(ctx context.Context, displayName string) (bool, error) {
	_, err := tc.ts.Get(ctx, displayName)
	if err == nil {
		return true, nil
	}

	if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
		return false, nil
	}

	return false, err
}

func extractDescription(teamData TeamData) string {
	if desc, ok := teamData["description"]; ok {
		if s, ok := desc.(string); ok {
			return s
		}
	}
	return ""
}

func extractStringSlice(teamData TeamData, key string) []string {
	if val, ok := teamData[key]; ok {
		if slice, ok := val.([]any); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return []string{}
}

func logExecutionResult(result *TeamCreateResult) {
	logger := initializers.Logger
	switch result.Status {
	case corecreator.StatusCreated:
		logger.Info("Team created successfully", "team", result.TeamName, "team_id", result.TeamID, "status", result.Status, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case corecreator.StatusAlreadyExists:
		logger.Info("Team already exists", "team", result.TeamName, "status", result.Status)
	case corecreator.StatusFailed:
		logger.Error("Team operation failed", "team", result.TeamName, "error", result.Error, "status", result.Status)
	}
}

func logDryRunResult(result *TeamCreateResult) {
	logger := initializers.Logger
	switch result.Status {
	case corecreator.StatusWouldCreate:
		logger.Info("Dry run: Team would be created", "team", result.TeamName, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case corecreator.StatusAlreadyExists:
		logger.Info("Dry run: Team already exists", "team", result.TeamName)
	case corecreator.StatusFailed:
		logger.Error("Dry run: Team creation would fail", "team", result.TeamName, "error", result.Error)
	}
}

func failedResult(err error, body *createTeamBody) *TeamCreateResult {
	return &TeamCreateResult{
		TeamName:    body.DisplayName,
		TeamID:      "",
		Error:       err,
		Status:      corecreator.StatusFailed,
		Description: body.Description,
	}
}

func alreadyExistsResult(body *createTeamBody) *TeamCreateResult {
	return &TeamCreateResult{
		TeamName:    body.DisplayName,
		TeamID:      "",
		Error:       nil,
		Status:      corecreator.StatusAlreadyExists,
		Description: body.Description,
	}
}

func staticAction(body *createTeamBody, result *TeamCreateResult) action {
	return corecreator.StaticAction(*body, *result)
}
