package utils

import (
	"strings"

	"github.com/pzsp-teams/lib/chats"
)

// GetChatRef returns a ChatRef for either group chat or
// one on one chat based on if given ref is an email
func GetChatRef(ref string) chats.ChatRef {
	if IsLikelyEmail(ref) || IsLikelyOneOnOneChatID(ref) {
		return chats.OneOnOneChatRef{Ref: ref}
	}
	if strings.TrimSpace(ref) == "" {
		return nil
	}
	return chats.GroupChatRef{Ref: ref}
}
