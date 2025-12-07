package utils

import (
	"errors"
)

var (
	ErrListingTeamsFailed    = errors.New("listing teams failed")
	ErrListingChannelsFailed = errors.New("listing channels failed")
	ErrListingMessagesFailed = errors.New("listing messages failed")
)
