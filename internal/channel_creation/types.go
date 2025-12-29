package channelcreation

import "context"

type createChannelBody struct {
	TeamRef    string
	ChannelRef string
	MemberRefs []string
	OwnerRefs  []string
}

// Status represents the result status of a channel creation attempt
type Status string
const (
	StatusCreated  Status = "Created"
	StatusAlreadyExists Status = "AlreadyExists"
	StatusMembersEnsured Status = "MembersEnsured"
	StatusFailed   Status = "Failed"
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

type EnsureMembersResult struct {
	MembersRefsEnsured []string
	OwnerRefsEnsured   []string
	MembersRefsFailed   []string
	OwnerRefsFailed     []string
}

type action struct {
	createChannelBody
	run func(ctx context.Context, body createChannelBody) error
}
