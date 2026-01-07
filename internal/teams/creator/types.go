package creator

import (
	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
)

// TeamData represents raw team data parsed from input files (YAML/JSON/CSV).
type TeamData struct {
	Description string   `yaml:"description" json:"description" toml:"description"`
	Owners      []string `yaml:"owners" json:"owners" toml:"owners"`
	Members     []string `yaml:"members" json:"members" toml:"members"`
	Visibility  string   `yaml:"visibility" json:"visibility" toml:"visibility"`
	IncludeMe   bool     `yaml:"includeMe" json:"includeMe" toml:"includeMe"`
}

type createTeamBody struct {
	DisplayName string
	Description string
	OwnerRefs   []string
	MemberRefs  []string
	Visibility  string
	IncludeMe   bool
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
	Visibility  string
}

type action = corecreator.Action[createTeamBody, TeamCreateResult]
