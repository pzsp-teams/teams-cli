package utils

import (
	"regexp"
	"strings"
)

// IsLikelyEmail returns true if given string matches email regexp
func IsLikelyEmail(s string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(s))
}
