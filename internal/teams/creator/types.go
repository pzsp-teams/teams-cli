package creator

import (
	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
)

// TeamData represents raw team data parsed from input files (YAML/JSON/CSV).
// Map keys include "description", "owners", and "members".
type TeamData map[string]any

type createTeamBody struct {
	DisplayName string
	Description string
	OwnerRefs   []string
	MemberRefs  []string
}

// TeamCreateResult represents the outcome of a team creation attempt.
type TeamCreateResult struct {
	TeamName    string
	TeamID      string
	Error       error
	Status      corecreator.Status
	MemberRefs  []string
	OwnerRefs   []string
	Description string
}

type action = corecreator.Action[createTeamBody, TeamCreateResult]
