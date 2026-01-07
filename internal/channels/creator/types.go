package creator

import (
	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
)

// ChannelData represents lists of members and owners for new channels
type ChannelData map[string][]string

type createChannelBody struct {
	TeamRef    string
	ChannelRef string
	MemberRefs []string
	OwnerRefs  []string
}

// Status represents the result status of a channel creation attempt
type Status = corecreator.Status

// Possible statuses for channel operations
const (
	StatusCreated            = corecreator.StatusCreated
	StatusWouldCreate        = corecreator.StatusWouldCreate
	StatusAlreadyExists      = corecreator.StatusAlreadyExists
	StatusMembersEnsured     = corecreator.StatusMembersEnsured
	StatusWouldEnsureMembers = corecreator.StatusWouldEnsureMembers
	StatusFailed             = corecreator.StatusFailed
	StatusPartiallyEnsured   = corecreator.StatusPartiallyEnsured
)

// CreateResult includes the result of a channel creation attempt
type CreateResult struct {
	ChannelName string
	ChannelID   string
	Error       error
	Status      Status
	MemberRefs  []string
	OwnerRefs   []string
}

type ensureMembersResult struct {
	MembersRefsEnsured []string
	OwnerRefsEnsured   []string
	MembersRefsFailed  []string
	OwnerRefsFailed    []string
}

type action = corecreator.Action[createChannelBody, CreateResult]
