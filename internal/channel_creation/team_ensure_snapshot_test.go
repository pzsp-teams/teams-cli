package channelcreation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_teamEnsureSnapshot_hasFailuresForBody(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")

	type tc struct {
		name     string
		snap     *teamEnsureSnapshot
		body     *createChannelBody
		wantHas  bool
		wantRefs []string
	}

	tests := []tc{
		{
			name:     "nil snapshot -> no failures",
			snap:     nil,
			body:     &createChannelBody{MemberRefs: []string{"u1"}, OwnerRefs: []string{"o1"}},
			wantHas:  false,
			wantRefs: nil,
		},
		{
			name:     "empty failed map -> no failures",
			snap:     &teamEnsureSnapshot{failed: map[string]error{}},
			body:     &createChannelBody{MemberRefs: []string{"u1"}, OwnerRefs: []string{"o1"}},
			wantHas:  false,
			wantRefs: nil,
		},
		{
			name:     "no matching refs in failed -> no failures",
			snap:     &teamEnsureSnapshot{failed: map[string]error{"uX": errBoom}},
			body:     &createChannelBody{MemberRefs: []string{"u1"}, OwnerRefs: []string{"o1"}},
			wantHas:  false,
			wantRefs: nil,
		},
		{
			name:     "member match -> returns that member",
			snap:     &teamEnsureSnapshot{failed: map[string]error{"u1": errBoom}},
			body:     &createChannelBody{MemberRefs: []string{"u1", "u2"}, OwnerRefs: []string{"o1"}},
			wantHas:  true,
			wantRefs: []string{"u1"},
		},
		{
			name:     "owner match -> returns that owner",
			snap:     &teamEnsureSnapshot{failed: map[string]error{"o1": errBoom}},
			body:     &createChannelBody{MemberRefs: []string{"u2"}, OwnerRefs: []string{"o1", "o2"}},
			wantHas:  true,
			wantRefs: []string{"o1"},
		},
		{
			name:     "both member and owner match -> returns both in body order (members then owners)",
			snap:     &teamEnsureSnapshot{failed: map[string]error{"u1": errBoom, "o2": errBoom}},
			body:     &createChannelBody{MemberRefs: []string{"u1", "uX"}, OwnerRefs: []string{"o1", "o2"}},
			wantHas:  true,
			wantRefs: []string{"u1", "o2"},
		},
		{
			name: "duplicates in body are deduplicated in output (first occurrence order kept)",
			snap: &teamEnsureSnapshot{failed: map[string]error{"u1": errBoom, "o1": errBoom}},
			body: &createChannelBody{
				MemberRefs: []string{"u1", "u1", "u2", "u1"},
				OwnerRefs:  []string{"o1", "o1"},
			},
			wantHas:  true,
			wantRefs: []string{"u1", "o1"},
		},
		{
			name: "same ref appears in members and owners -> appears once",
			snap: &teamEnsureSnapshot{failed: map[string]error{"x": errBoom}},
			body: &createChannelBody{
				MemberRefs: []string{"x"},
				OwnerRefs:  []string{"x"},
			},
			wantHas:  true,
			wantRefs: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			has, refs := tt.snap.hasFailuresForBody(tt.body)
			assert.Equal(t, tt.wantHas, has)
			assert.Equal(t, tt.wantRefs, refs)
		})
	}
}

func Test_uniqueNonEmpty(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string
		in   []string
		want []string
	}

	tests := []tc{
		{
			name: "nil input -> empty output",
			in:   nil,
			want: []string{},
		},
		{
			name: "empty slice -> empty output",
			in:   []string{},
			want: []string{},
		},
		{
			name: "trims spaces and removes empty",
			in:   []string{"  ", "\t", "\n", " u1 ", "", "u2"},
			want: []string{"u1", "u2"},
		},
		{
			name: "deduplicates and keeps first occurrence order after trimming",
			in:   []string{" u1 ", "u2", "u1", "u2 ", "u3", "u1"},
			want: []string{"u1", "u2", "u3"},
		},
		{
			name: "distinct values with spaces become same after trim",
			in:   []string{"a", " a ", "a  ", "b", " b"},
			want: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := uniqueNonEmpty(tt.in)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_shouldSkip(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")

	type tc struct {
		name string
		snap *teamEnsureSnapshot
		ref  string
		want bool
	}

	tests := []tc{
		{
			name: "nil snapshot -> false",
			snap: nil,
			ref:  "u1",
			want: false,
		},
		{
			name: "empty failed map -> false",
			snap: &teamEnsureSnapshot{failed: map[string]error{}},
			ref:  "u1",
			want: false,
		},
		{
			name: "ref not in failed -> false",
			snap: &teamEnsureSnapshot{failed: map[string]error{"u2": errBoom}},
			ref:  "u1",
			want: false,
		},
		{
			name: "ref in failed -> true",
			snap: &teamEnsureSnapshot{failed: map[string]error{"u1": errBoom}},
			ref:  "u1",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldSkip(tt.snap, tt.ref)
			assert.Equal(t, tt.want, got)
		})
	}
}
