package utils

import (
	"github.com/pzsp-teams/lib/chats"
)

// GetChatRef returns a ChatRef for either group chat or
// one on one chat based on if given ref is an email
func GetChatRef(ref string) chats.ChatRef {
	if IsLikelyEmail(ref) || IsLikelyOneOnOneChatID(ref) {
		return chats.OneOnOneChatRef{Ref: ref}
	}
	return chats.GroupChatRef{Ref: ref}
}
