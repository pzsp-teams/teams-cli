package retriever

import (
	"errors"
)

var (
	// ErrListingTeamsFailed indicates that listing teams failed
	ErrListingTeamsFailed = errors.New("listing teams failed")
	// ErrListingChannelsFailed indicates that listing channels failed
	ErrListingChannelsFailed = errors.New("listing channels failed")
	// ErrListingMessagesFailed indicates that listing messages failed
	ErrListingMessagesFailed = errors.New("listing messages failed")
	// ErrNoTeamsFound indicates that no teams were found
	ErrNoTeamsFound = errors.New("no teams found")
	// ErrNoChannelsFound indicates that no channels were found
	ErrNoChannelsFound = errors.New("no channels found")
)
