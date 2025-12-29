package channelcreation

import "errors"

var (
	errDataParseFailed = errors.New("failed to parse channels data")
	errMembersPartiallyEnsured = errors.New("some members or owners could not be ensured in the channel")
)
