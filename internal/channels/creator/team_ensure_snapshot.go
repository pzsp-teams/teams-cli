package creator

type teamEnsureSnapshot struct {
	planned []string
	ensured map[string]struct{}
	failed  map[string]error
}

func (s *teamEnsureSnapshot) hasFailuresForBody(body *createChannelBody) (has bool, failedRefs []string) {
	if s == nil || len(s.failed) == 0 {
		return false, nil
	}
	seen := map[string]struct{}{}
	for _, ref := range append(body.MemberRefs, body.OwnerRefs...) {
		if _, ok := s.failed[ref]; ok {
			if _, alreadySeen := seen[ref]; !alreadySeen {
				failedRefs = append(failedRefs, ref)
				seen[ref] = struct{}{}
			}
		}
	}
	return len(failedRefs) > 0, failedRefs
}

func shouldSkip(t *teamEnsureSnapshot, ref string) bool {
	if t == nil || len(t.failed) == 0 {
		return false
	}
	_, ok := t.failed[ref]
	return ok
}
