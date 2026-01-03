package channelcreation

import "strings"

type teamEnsureSnapshot struct {
	planned []string
	ensured map[string]struct{}
	failed  map[string]error
}

func (s *teamEnsureSnapshot) hasFailuresForBody(body *createChannelBody) (bool, []string) {
	if s == nil || len(s.failed) == 0 {
		return false, nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, ref := range append(body.MemberRefs, body.OwnerRefs...) {
		if _, ok := s.failed[ref]; ok {
			if _, alreadySeen := seen[ref]; !alreadySeen {
				out = append(out, ref)
				seen[ref] = struct{}{}
			}
		}
	}
	return len(out) > 0, out
}

func uniqueNonEmpty(refs []string) []string {
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

func shouldSkip(t *teamEnsureSnapshot, ref string) bool {
	if t == nil || len(t.failed) == 0 {
		return false
	}
	_, ok := t.failed[ref]
	return ok
}
