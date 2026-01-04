package retriever

import "errors"

var (
	// ErrListingMessagesFailed indicates that listing chat messages failed
	ErrListingMessagesFailed = errors.New("listing chat messages failed")
)
