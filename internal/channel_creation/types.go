package channelcreation

import "context"

type ChannelData map[string][]string

type createChannelBody struct {
	TeamRef    string
	ChannelRef string
	MemberRefs []string
	OwnerRefs  []string
}

// Status represents the result status of a channel creation attempt
type Status string

// Possible statuses for channel operations
const (
	StatusCreated            Status = "Created"
	StatusWouldCreate        Status = "WouldCreate"
	StatusAlreadyExists      Status = "AlreadyExists"
	StatusMembersEnsured     Status = "MembersEnsured"
	StatusWouldEnsureMembers Status = "WouldEnsureMembers"
	StatusFailed             Status = "Failed"
	StatusPartiallyEnsured   Status = "PartiallyEnsured"
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

type action struct {
	createChannelBody
	run    func(ctx context.Context, body createChannelBody) *CreateResult
	result *CreateResult
}
