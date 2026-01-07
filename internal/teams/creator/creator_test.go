package creator

import (
	"context"
	"errors"
	"sort"
	"testing"

	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newMocks(t *testing.T) (*gomock.Controller, *testutil.MockTeamsService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	ts := testutil.NewMockTeamsService(ctrl)
	return ctrl, ts
}

func resultsByName(results []TeamCreateResult) map[string]TeamCreateResult {
	out := make(map[string]TeamCreateResult, len(results))
	for i := range results {
		r := results[i]
		out[r.TeamName] = r
	}
	return out
}

func sortedKeys(m map[string]TeamCreateResult) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var errBoom = errors.New("boom")

type wantResult struct {
	status      corecreator.Status
	teamID      string
	errNil      bool
	errIs       error
	errContains string
	members     []string
	owners      []string
	description string
	visibility  string
}

func assertResult(t *testing.T, gotMap map[string]TeamCreateResult, name string, w *wantResult) {
	t.Helper()

	r, ok := gotMap[name]
	require.True(t, ok, "missing result for team %q; got keys=%v", name, sortedKeys(gotMap))

	assert.Equal(t, w.status, r.Status)
	assert.Equal(t, w.teamID, r.TeamID)

	if w.errNil {
		assert.NoError(t, r.Error)
	} else {
		assert.Error(t, r.Error)
		if w.errIs != nil {
			assert.ErrorIs(t, r.Error, w.errIs)
		}
		if w.errContains != "" {
			assert.Contains(t, r.Error.Error(), w.errContains)
		}
	}

	if w.members != nil {
		assert.Equal(t, w.members, r.MemberRefs)
	}
	if w.owners != nil {
		assert.Equal(t, w.owners, r.OwnerRefs)
	}
	if w.description != "" {
		assert.Equal(t, w.description, r.Description)
	}
	if w.visibility != "" {
		assert.Equal(t, w.visibility, r.Visibility)
	}
}

func TestTeamCreator_Create_dryRun_teamMissing_wouldCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			Description: "Alpha team description",
			Members:     []string{"u1"},
			Owners:      []string{"o1"},
		},
	}, true)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusWouldCreate,
		errNil:      true,
		members:     []string{"u1"},
		owners:      []string{"o1"},
		description: "Alpha team description",
		visibility:  "private",
	})
}

func TestTeamCreator_Create_teamMissing_createsTeam(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha description", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-id-123", nil).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			Description: "Alpha description",
			Members:     []string{"u1"},
			Owners:      []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-id-123",
		errNil:      true,
		members:     []string{"u1"},
		owners:      []string{"o1"},
		description: "Alpha description",
		visibility:  "private",
	})
}

func TestTeamCreator_Create_createFromTemplateFails_returnsFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha description", []string{"o1"}, []string{"u1"}, "private", false).
		Return("", errBoom).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			Description: "Alpha description",
			Members:     []string{"u1"},
			Owners:      []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusFailed,
		errNil:      false,
		errIs:       errBoom,
		errContains: "failed to create team",
		description: "Alpha description",
		visibility:  "private",
	})
}

func TestTeamCreator_Create_multipleTeams_processesAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha desc", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-alpha-id", nil)

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Gamma", "Gamma desc", []string{"o3"}, []string{}, "private", false).
		Return("team-gamma-id", nil)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			Description: "Alpha desc",
			Members:     []string{"u1"},
			Owners:      []string{"o1"},
		},
		"Team Gamma": {
			Description: "Gamma desc",
			Members:     []string{},
			Owners:      []string{"o3"},
		},
	}, false)

	require.Len(t, got, 2)
	gotMap := resultsByName(got)

	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-alpha-id",
		errNil:      true,
		members:     []string{"u1"},
		owners:      []string{"o1"},
		description: "Alpha desc",
		visibility:  "private",
	})

	assertResult(t, gotMap, "Team Gamma", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-gamma-id",
		errNil:      true,
		members:     []string{},
		owners:      []string{"o3"},
		description: "Gamma desc",
		visibility:  "private",
	})
}

func TestTeamCreator_Create_emptyRequest_returnsEmptyResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{}, false)

	require.Len(t, got, 0)
}

func TestTeamCreator_Create_teamWithoutDescription_usesEmptyString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-id", nil)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			Members: []string{"u1"},
			Owners:  []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-id",
		errNil:      true,
		description: "",
		visibility:  "private",
	})
}
