package retriever

import "errors"

var (
	// ErrListingChatsFailed indicates that listing chats failed
	ErrListingChatsFailed = errors.New("listing chats failed")
	// ErrNoChatsFound indicates that no chats were found
	ErrNoChatsFound = errors.New("no chats found")
	// ErrListingMessagesFailed indicates that listing chat messages failed
	ErrListingMessagesFailed = errors.New("listing chat messages failed")
)
