package utils

import (
	"regexp"
	"strings"
)

// IsLikelyGroupChatID checks if the given string is likely a group chat ID
func IsLikelyGroupChatID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@thread.")
}

// IsLikelyOneOnOneChatID checks if the given string is likely a one-on-one chat ID
func IsLikelyOneOnOneChatID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@unq.")
}

// IsLikelyGUID checks if the given string is likely a GUID
func IsLikelyGUID(s string) bool {
	s = strings.TrimSpace(s)
	var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return guidRegex.MatchString(s)
}

// IsLikelyEmail checks if the given string is likely an email address
func IsLikelyEmail(s string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(s))
}
