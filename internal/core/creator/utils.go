package creator

import (
	"slices"
	"strings"
)

// UniqueNonEmpty returns a deduplicated slice of non-empty, trimmed strings
// preserving the order of first occurrence. Empty and whitespace-only strings
// are filtered out.
func UniqueNonEmpty(refs []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; !ok {
			out = append(out, ref)
			seen[ref] = struct{}{}
		}
	}
	return out
}

// Contains checks if a string slice contains a specific item.
func Contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
