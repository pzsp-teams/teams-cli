package messaging

import "errors"

var (
	// ErrMessageSkipped marks a message as skipped in case ignoreError is not set when sending messages
	ErrMessageSkipped = errors.New("skipped sending message, earlier message failed")

	// ErrMentionResolutionFailed indicates that mention resolution failed
	ErrMentionResolutionFailed = errors.New("failed to resolve mentions")

	// ErrMessageSendFailed indicates that sending a message failed
	ErrMessageSendFailed = errors.New("failed to send message")
)
