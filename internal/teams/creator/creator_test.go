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

var (
	errNotFound = errors.New("[CODE: 404] not found")
	errBoom     = errors.New("boom")
)

type wantResult struct {
	status      corecreator.Status
	teamID      string
	errNil      bool
	errIs       error
	errContains string
	members     []string
	owners      []string
	description string
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
}

func TestTeamCreator_Create_dryRun_teamMissing_wouldCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, errNotFound).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha team description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
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
	})
}

func TestTeamCreator_Create_teamMissing_createsTeam(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, errNotFound).
		Times(1)

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha description", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-id-123", nil).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
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
	})
}

func TestTeamCreator_Create_teamExists_skipsCreation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, nil).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusAlreadyExists,
		errNil:      true,
		description: "Alpha description",
	})
}

func TestTeamCreator_Create_dryRun_teamExists_returnsAlreadyExists(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, nil).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
		},
	}, true)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status: corecreator.StatusAlreadyExists,
		errNil: true,
	})
}

func TestTeamCreator_Create_existenceCheckFails_returnsFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, errBoom).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusFailed,
		errNil:      false,
		errIs:       errBoom,
		errContains: "failed to check existence",
	})
}

func TestTeamCreator_Create_createFromTemplateFails_returnsFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().
		Get(gomock.Any(), "Team Alpha").
		Return(nil, errNotFound).
		Times(1)

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha description", []string{"o1"}, []string{"u1"}, "private", false).
		Return("", errBoom).
		Times(1)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha description",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusFailed,
		errNil:      false,
		errIs:       errBoom,
		errContains: "failed to create team",
	})
}

func TestTeamCreator_Create_multipleTeams_processesAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().Get(gomock.Any(), "Team Alpha").Return(nil, errNotFound)
	ts.EXPECT().Get(gomock.Any(), "Team Beta").Return(nil, nil)
	ts.EXPECT().Get(gomock.Any(), "Team Gamma").Return(nil, errNotFound)

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "Alpha desc", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-alpha-id", nil)

	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Gamma", "Gamma desc", []string{"o3"}, []string{}, "private", false).
		Return("team-gamma-id", nil)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"description": "Alpha desc",
			"members":     []string{"u1"},
			"owners":      []string{"o1"},
		},
		"Team Beta": {
			"description": "Beta desc",
			"members":     []string{"u2"},
			"owners":      []string{"o2"},
		},
		"Team Gamma": {
			"description": "Gamma desc",
			"members":     []string{},
			"owners":      []string{"o3"},
		},
	}, false)

	require.Len(t, got, 3)
	gotMap := resultsByName(got)

	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-alpha-id",
		errNil:      true,
		members:     []string{"u1"},
		owners:      []string{"o1"},
		description: "Alpha desc",
	})

	assertResult(t, gotMap, "Team Beta", &wantResult{
		status: corecreator.StatusAlreadyExists,
		errNil: true,
	})

	assertResult(t, gotMap, "Team Gamma", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-gamma-id",
		errNil:      true,
		members:     []string{},
		owners:      []string{"o3"},
		description: "Gamma desc",
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

	ts.EXPECT().Get(gomock.Any(), "Team Alpha").Return(nil, errNotFound)
	ts.EXPECT().
		CreateFromTemplate(gomock.Any(), "Team Alpha", "", []string{"o1"}, []string{"u1"}, "private", false).
		Return("team-id", nil)

	sut := NewTeamCreator(ts)
	got := sut.Create(ctx, map[string]TeamData{
		"Team Alpha": {
			"members": []string{"u1"},
			"owners":  []string{"o1"},
		},
	}, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "Team Alpha", &wantResult{
		status:      corecreator.StatusCreated,
		teamID:      "team-id",
		errNil:      true,
		description: "",
	})
}

func TestExtractDescription_returnsStringValue(t *testing.T) {
	teamData := TeamData{"description": "My description"}
	got := extractDescription(teamData)
	require.Equal(t, "My description", got)
}

func TestExtractDescription_missingKey_returnsEmpty(t *testing.T) {
	teamData := TeamData{"other": "value"}
	got := extractDescription(teamData)
	require.Equal(t, "", got)
}

func TestExtractDescription_nonStringValue_returnsEmpty(t *testing.T) {
	teamData := TeamData{"description": 123}
	got := extractDescription(teamData)
	require.Equal(t, "", got)
}

func TestExtractStringSlice_interfaceSlice(t *testing.T) {
	teamData := TeamData{"members": []any{"u1", "u2"}}
	got := extractStringSlice(teamData, "members")
	require.Equal(t, []string{"u1", "u2"}, got)
}

func TestExtractStringSlice_stringSlice(t *testing.T) {
	teamData := TeamData{"members": []string{"u1", "u2"}}
	got := extractStringSlice(teamData, "members")
	require.Equal(t, []string{"u1", "u2"}, got)
}

func TestExtractStringSlice_missingKey_returnsEmpty(t *testing.T) {
	teamData := TeamData{"other": []string{"u1"}}
	got := extractStringSlice(teamData, "members")
	require.Equal(t, []string{}, got)
}

func TestExtractStringSlice_nonSliceValue_returnsEmpty(t *testing.T) {
	teamData := TeamData{"members": "not a slice"}
	got := extractStringSlice(teamData, "members")
	require.Equal(t, []string{}, got)
}
