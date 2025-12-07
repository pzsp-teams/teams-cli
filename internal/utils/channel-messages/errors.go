package utils

import (
	"errors"
)

// ErrListingTeamsFailed will be used later
var (
	ErrListingTeamsFailed    = errors.New("listing teams failed")
	ErrListingChannelsFailed = errors.New("listing channels failed")
	ErrListingMessagesFailed = errors.New("listing messages failed")
	ErrNoTeamsFound          = errors.New("no teams found")
	ErrNoChannelsFound       = errors.New("no channels found")
)
