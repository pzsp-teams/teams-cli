package creator

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newMocks(t *testing.T) (*gomock.Controller, *testutil.MockChannelsService, *testutil.MockTeamsService) {
	t.Helper()
	ctrl := gomock.NewController(t)

	ch := testutil.NewMockChannelsService(ctrl)
	ts := testutil.NewMockTeamsService(ctrl)

	return ctrl, ch, ts
}

func newMocksNoTeams(t *testing.T) (*gomock.Controller, *testutil.MockChannelsService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	ch := testutil.NewMockChannelsService(ctrl)
	return ctrl, ch
}

func resultsByName(results []CreateResult) map[string]CreateResult {
	out := make(map[string]CreateResult, len(results))
	for _, r := range results {
		out[r.ChannelName] = r
	}
	return out
}

func sortedKeys(m map[string]CreateResult) []string {
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
	status      Status
	channelID   string
	errNil      bool
	errIs       error
	errContains string
	members     []string
	owners      []string
}

func assertResult(t *testing.T, gotMap map[string]CreateResult, name string, w *wantResult) {
	t.Helper()

	r, ok := gotMap[name]
	require.True(t, ok, "missing result for channel %q; got keys=%v", name, sortedKeys(gotMap))

	assert.Equal(t, w.status, r.Status)
	assert.Equal(t, w.channelID, r.ChannelID)

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
}

func TestChannelCreator_Create_requiresTeamsServiceWhenEnsureMembersInTeam(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch := newMocksNoTeams(t)
	defer ctrl.Finish()

	sut := NewChannelCreator(ch, nil)

	req := map[string]ChannelData{
		"Chan1": {"members": []string{"u1"}, "owners": []string{"o1"}},
		"Chan2": {"members": []string{"u2"}, "owners": []string{}},
	}

	got := sut.Create(ctx, teamRef, req, false, true, false)

	require.Len(t, got, len(req))
	gotMap := resultsByName(got)

	assertResult(t, gotMap, "Chan1", &wantResult{status: StatusFailed, errNil: false, errContains: "requires teams service"})
	assertResult(t, gotMap, "Chan2", &wantResult{status: StatusFailed, errNil: false, errContains: "requires teams service"})
}

func TestChannelCreator_Create_dryRun_channelMissing_wouldCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(nil, errNotFound).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1"}, "owners": []string{"o1"}},
	}, false, false, true)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:  StatusWouldCreate,
		errNil:  true,
		members: []string{"u1"},
		owners:  []string{"o1"},
	})
}

func TestChannelCreator_Create_dryRun_ensureMembersInTeam_doesNotCallTeamsAddMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(nil, errNotFound).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1"}, "owners": []string{"o1"}},
	}, false, true, true)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{status: StatusWouldCreate, errNil: true})
}

func TestChannelCreator_Create_channelMissing_createsChannel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(nil, errNotFound).
		Times(1)

	ch.EXPECT().
		CreatePrivateChannel(gomock.Any(), teamRef, "ChanA", []string{"u1"}, []string{"o1"}).
		Return(&models.Channel{ID: "chan-id"}, nil).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1"}, "owners": []string{"o1"}},
	}, false, false, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:    StatusCreated,
		channelID: "chan-id",
		errNil:    true,
		members:   []string{"u1"},
		owners:    []string{"o1"},
	})
}

func TestChannelCreator_Create_channelMissing_createFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(nil, errNotFound).
		Times(1)

	ch.EXPECT().
		CreatePrivateChannel(gomock.Any(), teamRef, "ChanA", []string{"u1"}, []string{"o1"}).
		Return(nil, errBoom).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1"}, "owners": []string{"o1"}},
	}, false, false, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:      StatusFailed,
		errNil:      false,
		errContains: "failed to create channel ChanA in team TeamA",
	})
}

func TestChannelCreator_Create_channelExists_alreadyExistsWhenEnsureMembersFalse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(&models.Channel{ID: "existing"}, nil).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1"}, "owners": []string{"o1"}},
	}, false, false, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{status: StatusAlreadyExists, errNil: true})
}

func TestChannelCreator_Create_channelExists_ensureMembers_success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(&models.Channel{ID: "existing"}, nil).
		Times(1)

	ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "u1", false).Return(nil, nil).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "u2", false).Return(nil, nil).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "o1", true).Return(nil, nil).Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"u1", "u2"}, "owners": []string{"o1"}},
	}, true, false, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:  StatusMembersEnsured,
		errNil:  true,
		members: []string{"u1", "u2"},
		owners:  []string{"o1"},
	})
}

func TestChannelCreator_Create_channelExists_ensureMembers_partialCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	tests := []struct {
		name       string
		req        map[string]ChannelData
		setupMocks func(ch *testutil.MockChannelsService)
		want       wantResult
	}{
		{
			name: "member add fails => partially ensured",
			req: map[string]ChannelData{
				"ChanA": {"members": []string{"uBad", "uOK"}, "owners": []string{"oOK"}},
			},
			setupMocks: func(ch *testutil.MockChannelsService) {
				ch.EXPECT().Get(gomock.Any(), teamRef, "ChanA").Return(&models.Channel{ID: "existing"}, nil).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "uBad", false).Return(nil, errBoom).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "uOK", false).Return(nil, nil).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "oOK", true).Return(nil, nil).Times(1)
			},
			want: wantResult{
				status:  StatusPartiallyEnsured,
				errNil:  false,
				errIs:   errMembersPartiallyEnsured,
				members: []string{"uOK"},
				owners:  []string{"oOK"},
			},
		},
		{
			name: "owner add fails => partially ensured",
			req: map[string]ChannelData{
				"ChanA": {"members": []string{"uOK"}, "owners": []string{"oBad", "oOK"}},
			},
			setupMocks: func(ch *testutil.MockChannelsService) {
				ch.EXPECT().Get(gomock.Any(), teamRef, "ChanA").Return(&models.Channel{ID: "existing"}, nil).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "uOK", false).Return(nil, nil).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "oBad", true).Return(nil, errBoom).Times(1)
				ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "oOK", true).Return(nil, nil).Times(1)
			},
			want: wantResult{
				status:  StatusPartiallyEnsured,
				errNil:  false,
				errIs:   errMembersPartiallyEnsured,
				members: []string{"uOK"},
				owners:  []string{"oOK"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl, ch, ts := newMocks(t)
			defer ctrl.Finish()
			_ = ts

			tt.setupMocks(ch)

			sut := NewChannelCreator(ch, ts)
			got := sut.Create(ctx, teamRef, tt.req, true, false, false)

			require.Len(t, got, 1)
			gotMap := resultsByName(got)
			assertResult(t, gotMap, "ChanA", &tt.want)
		})
	}
}

func TestChannelCreator_Create_ensureMembersInTeam_failureBlocksCreateForMissingChannel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().AddMember(gomock.Any(), teamRef, "uFail", false).Return(nil, errBoom).Times(1)

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(nil, errNotFound).
		Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"uFail"}, "owners": []string{}},
	}, false, true, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:      StatusFailed,
		errNil:      false,
		errContains: "cannot create channel",
	})
}

func TestChannelCreator_Create_ensureMembersInTeam_failureSkipsEnsureInChannel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().AddMember(gomock.Any(), teamRef, "uFail", false).Return(nil, errBoom).Times(1)
	ts.EXPECT().AddMember(gomock.Any(), teamRef, "uOK", false).Return(nil, nil).Times(1)
	ts.EXPECT().AddMember(gomock.Any(), teamRef, "oOK", false).Return(nil, nil).Times(1)

	ch.EXPECT().
		Get(gomock.Any(), teamRef, "ChanA").
		Return(&models.Channel{ID: "existing"}, nil).
		Times(1)

	ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "uOK", false).Return(nil, nil).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), teamRef, "ChanA", "oOK", true).Return(nil, nil).Times(1)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"ChanA": {"members": []string{"uFail", "uOK"}, "owners": []string{"oOK"}},
	}, true, true, false)

	require.Len(t, got, 1)
	gotMap := resultsByName(got)
	assertResult(t, gotMap, "ChanA", &wantResult{
		status:  StatusPartiallyEnsured,
		errNil:  false,
		errIs:   errMembersPartiallyEnsured,
		members: []string{"uOK"},
		owners:  []string{"oOK"},
	})
}

func TestChannelCreator_Create_ensureMembersInTeam_deduplicatesRefsAcrossBodies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const teamRef = "TeamA"

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()

	ts.EXPECT().AddMember(gomock.Any(), teamRef, "u1", false).Return(nil, nil).Times(1)

	ch.EXPECT().Get(gomock.Any(), teamRef, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, channelRef string) (*models.Channel, error) {
			switch channelRef {
			case "Chan1", "Chan2":
				return nil, errNotFound
			default:
				return nil, errors.New("unexpected channelRef: " + channelRef)
			}
		}).
		Times(2)

	ch.EXPECT().CreatePrivateChannel(gomock.Any(), teamRef, gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, channelRef string, _ []string, _ []string) (*models.Channel, error) {
			return &models.Channel{ID: "id-" + channelRef}, nil
		}).
		Times(2)

	sut := NewChannelCreator(ch, ts)
	got := sut.Create(ctx, teamRef, map[string]ChannelData{
		"Chan1": {"members": []string{"u1", "u1", " "}, "owners": []string{}},
		"Chan2": {"members": []string{"u1"}, "owners": []string{""}},
	}, false, true, false)

	require.Len(t, got, 2)
	gotMap := resultsByName(got)

	assertResult(t, gotMap, "Chan1", &wantResult{status: StatusCreated, channelID: "id-Chan1", errNil: true})
	assertResult(t, gotMap, "Chan2", &wantResult{status: StatusCreated, channelID: "id-Chan2", errNil: true})
}

func TestChannelCreator_executeActions_handlesNilReturnedResult(t *testing.T) {
	t.Parallel()

	cc := &channelCreator{}
	actions := []*action{
		{
			createChannelBody: createChannelBody{
				TeamRef:    "TeamA",
				ChannelRef: "ChanA",
			},
			run: func(ctx context.Context, body createChannelBody) *CreateResult {
				return nil
			},
		},
	}

	got := cc.executeActions(context.Background(), actions)
	require.Len(t, got, 1)

	assert.Equal(t, "ChanA", got[0].ChannelName)
	assert.Equal(t, "", got[0].ChannelID)
	assert.Equal(t, StatusFailed, got[0].Status)
	require.Error(t, got[0].Error)
	assert.Contains(t, got[0].Error.Error(), "action returned nil result")
}

func TestChannelCreator_dryRunActions_handlesNilStaticResult(t *testing.T) {
	t.Parallel()

	cc := &channelCreator{}
	actions := []*action{
		{
			createChannelBody: createChannelBody{
				TeamRef:    "TeamA",
				ChannelRef: "ChanA",
			},
			result: nil,
			run: func(ctx context.Context, body createChannelBody) *CreateResult {
				return nil
			},
		},
	}

	got := cc.dryRunActions(actions)
	require.Len(t, got, 1)

	assert.Equal(t, "ChanA", got[0].ChannelName)
	assert.Equal(t, "", got[0].ChannelID)
	assert.Equal(t, StatusFailed, got[0].Status)
	require.Error(t, got[0].Error)
	assert.Contains(t, got[0].Error.Error(), "action has nil result in dry run")
}

func Test_failedResult(t *testing.T) {
	t.Parallel()

	body := &createChannelBody{
		TeamRef:    "TeamA",
		ChannelRef: "ChanA",
		MemberRefs: []string{"u1"},
		OwnerRefs:  []string{"o1"},
	}

	err := errors.New("x")
	got := failedResult(err, body)

	require.NotNil(t, got)
	assert.Equal(t, "ChanA", got.ChannelName)
	assert.Equal(t, "", got.ChannelID)
	assert.Equal(t, StatusFailed, got.Status)
	assert.Same(t, err, got.Error)
}

func Test_alreadyExistsResult(t *testing.T) {
	t.Parallel()

	body := &createChannelBody{
		TeamRef:    "TeamA",
		ChannelRef: "ChanA",
	}
	got := alreadyExistsResult(body)

	require.NotNil(t, got)
	assert.Equal(t, "ChanA", got.ChannelName)
	assert.Equal(t, "", got.ChannelID)
	assert.Equal(t, StatusAlreadyExists, got.Status)
	assert.NoError(t, got.Error)
}

func Test_staticAction(t *testing.T) {
	t.Parallel()

	body := &createChannelBody{
		TeamRef:    "TeamA",
		ChannelRef: "ChanA",
	}
	res := &CreateResult{
		ChannelName: "ChanA",
		Status:      StatusAlreadyExists,
	}

	act := staticAction(body, res)
	require.NotNil(t, act)
	require.NotNil(t, act.run)
	assert.Same(t, res, act.result)

	got := act.run(context.Background(), act.createChannelBody)
	assert.Same(t, res, got)
}

func Test_channelCreator_transformRequestToCreateChannelBody(t *testing.T) {
	t.Parallel()

	cc := &channelCreator{}
	req := map[string]ChannelData{
		"ChanA": {"members": []string{"u1", "u2"}, "owners": []string{"o1"}},
		"ChanB": {"members": []string{}, "owners": []string{"o2", "o3"}},
	}

	bodies := cc.transformRequestToCreateChannelBody("TeamA", req)
	require.Len(t, bodies, 2)

	gotByRef := map[string]createChannelBody{}
	for _, b := range bodies {
		gotByRef[b.ChannelRef] = b
	}

	require.Contains(t, gotByRef, "ChanA")
	assert.Equal(t, "TeamA", gotByRef["ChanA"].TeamRef)
	assert.Equal(t, []string{"u1", "u2"}, gotByRef["ChanA"].MemberRefs)
	assert.Equal(t, []string{"o1"}, gotByRef["ChanA"].OwnerRefs)

	require.Contains(t, gotByRef, "ChanB")
	assert.Equal(t, "TeamA", gotByRef["ChanB"].TeamRef)
	assert.Equal(t, []string{}, gotByRef["ChanB"].MemberRefs)
	assert.Equal(t, []string{"o2", "o3"}, gotByRef["ChanB"].OwnerRefs)
}

func Test_channelCreator_checkChannelExists(t *testing.T) {
	t.Parallel()

	type tc struct {
		name       string
		getErr     error
		wantExists bool
		wantErr    bool
	}

	tests := []tc{
		{name: "Get returns nil error -> exists", getErr: nil, wantExists: true, wantErr: false},
		{name: "Get returns [CODE: 404] -> does not exist, no error", getErr: errNotFound, wantExists: false, wantErr: false},
		{name: "Get returns 'not found' string -> does not exist, no error", getErr: errors.New("resource not found"), wantExists: false, wantErr: false},
		{name: "Get returns other error -> error", getErr: errBoom, wantExists: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl, ch, ts := newMocks(t)
			defer ctrl.Finish()
			_ = ts

			cc := &channelCreator{chans: ch}

			ch.EXPECT().
				Get(gomock.Any(), "TeamA", "ChanA").
				DoAndReturn(func(_ context.Context, _ string, _ string) (*models.Channel, error) {
					if tt.getErr == nil {
						return &models.Channel{ID: "id"}, nil
					}
					return nil, tt.getErr
				}).
				Times(1)

			exists, err := cc.checkChannelExists(context.Background(), "TeamA", "ChanA")
			assert.Equal(t, tt.wantExists, exists)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_channelCreator_ensureMembersInTeamSnapshot_dryRun(t *testing.T) {
	t.Parallel()

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ch

	cc := &channelCreator{ts: ts}

	bodies := []createChannelBody{
		{TeamRef: "TeamA", ChannelRef: "ChanA", MemberRefs: []string{"u1", "u1", " "}, OwnerRefs: []string{"o1", ""}},
		{TeamRef: "TeamA", ChannelRef: "ChanB", MemberRefs: []string{"u1"}, OwnerRefs: []string{"o1"}},
	}

	snap := cc.ensureMembersInTeamSnapshot(context.Background(), "TeamA", bodies, true)

	require.NotNil(t, snap)
	assert.ElementsMatch(t, []string{"u1", "o1"}, snap.planned)
	assert.Empty(t, snap.ensured)
	assert.Empty(t, snap.failed)
}

func Test_channelCreator_ensureMembersInTeamSnapshot_nonDryRun_collectsEnsuredAndFailed_andDedups(t *testing.T) {
	t.Parallel()

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ch

	cc := &channelCreator{ts: ts}

	bodies := []createChannelBody{
		{TeamRef: "TeamA", ChannelRef: "ChanA", MemberRefs: []string{"u1", "uBad", "u1"}, OwnerRefs: []string{"o1"}},
		{TeamRef: "TeamA", ChannelRef: "ChanB", MemberRefs: []string{"u1"}, OwnerRefs: []string{"o1", ""}},
	}

	ts.EXPECT().AddMember(gomock.Any(), "TeamA", "u1", false).Return(nil, nil).Times(1)
	ts.EXPECT().AddMember(gomock.Any(), "TeamA", "uBad", false).Return(nil, errBoom).Times(1)
	ts.EXPECT().AddMember(gomock.Any(), "TeamA", "o1", false).Return(nil, nil).Times(1)

	snap := cc.ensureMembersInTeamSnapshot(context.Background(), "TeamA", bodies, false)

	require.NotNil(t, snap)
	assert.ElementsMatch(t, []string{"u1", "uBad", "o1"}, snap.planned)

	_, ok := snap.ensured["u1"]
	assert.True(t, ok)
	_, ok = snap.ensured["o1"]
	assert.True(t, ok)

	require.Contains(t, snap.failed, "uBad")
	assert.ErrorIs(t, snap.failed["uBad"], errBoom)
}

func Test_channelCreator_ensureMembersInChannel_respectsSkip_andCollects(t *testing.T) {
	t.Parallel()

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	cc := &channelCreator{chans: ch}

	teamSnap := &teamEnsureSnapshot{
		failed: map[string]error{
			"uSkip": errBoom,
			"oSkip": errBoom,
		},
		ensured: map[string]struct{}{},
	}

	body := &createChannelBody{
		TeamRef:    "TeamA",
		ChannelRef: "ChanA",
		MemberRefs: []string{"uSkip", "uOK"},
		OwnerRefs:  []string{"oSkip", "oOK"},
	}

	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "uOK", false).Return(nil, nil).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "oOK", true).Return(nil, nil).Times(1)

	got := cc.ensureMembersInChannel(context.Background(), body, teamSnap)
	require.NotNil(t, got)

	assert.ElementsMatch(t, []string{"uSkip"}, got.MembersRefsFailed)
	assert.ElementsMatch(t, []string{"oSkip"}, got.OwnerRefsFailed)
	assert.ElementsMatch(t, []string{"uOK"}, got.MembersRefsEnsured)
	assert.ElementsMatch(t, []string{"oOK"}, got.OwnerRefsEnsured)
}

func Test_channelCreator_ensureMembersInChannel_recordsAddMemberErrors(t *testing.T) {
	t.Parallel()

	ctrl, ch, ts := newMocks(t)
	defer ctrl.Finish()
	_ = ts

	cc := &channelCreator{chans: ch}

	body := &createChannelBody{
		TeamRef:    "TeamA",
		ChannelRef: "ChanA",
		MemberRefs: []string{"uBad", "uOK"},
		OwnerRefs:  []string{"oBad", "oOK"},
	}

	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "uBad", false).Return(nil, errBoom).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "uOK", false).Return(nil, nil).Times(1)

	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "oBad", true).Return(nil, errBoom).Times(1)
	ch.EXPECT().AddMember(gomock.Any(), "TeamA", "ChanA", "oOK", true).Return(nil, nil).Times(1)

	got := cc.ensureMembersInChannel(context.Background(), body, nil)
	require.NotNil(t, got)

	assert.ElementsMatch(t, []string{"uBad"}, got.MembersRefsFailed)
	assert.ElementsMatch(t, []string{"oBad"}, got.OwnerRefsFailed)
	assert.ElementsMatch(t, []string{"uOK"}, got.MembersRefsEnsured)
	assert.ElementsMatch(t, []string{"oOK"}, got.OwnerRefsEnsured)
}

func Test_channelCreator_planActionForBody_branches(t *testing.T) {
	t.Parallel()

	type tc struct {
		name string

		existsErr     error
		ensureMembers bool
		teamSnap      *teamEnsureSnapshot

		wantStatus      Status
		wantErrContains string
	}

	tests := []tc{
		{
			name:          "exists and ensureMembers=false -> already exists action",
			existsErr:     nil,
			ensureMembers: false,
			teamSnap:      nil,
			wantStatus:    StatusAlreadyExists,
		},
		{
			name:          "exists and ensureMembers=true -> would ensure members action",
			existsErr:     nil,
			ensureMembers: true,
			teamSnap:      nil,
			wantStatus:    StatusWouldEnsureMembers,
		},
		{
			name:          "not exists and no teamSnap failure -> would create",
			existsErr:     errNotFound,
			ensureMembers: false,
			teamSnap:      nil,
			wantStatus:    StatusWouldCreate,
		},
		{
			name:          "not exists and teamSnap has failure for body -> failed",
			existsErr:     errNotFound,
			ensureMembers: false,
			teamSnap: &teamEnsureSnapshot{
				failed:  map[string]error{"uBad": errBoom},
				ensured: map[string]struct{}{},
			},
			wantStatus:      StatusFailed,
			wantErrContains: "cannot create channel",
		},
		{
			name:            "checkChannelExists returns non-404 error -> failed",
			existsErr:       errBoom,
			ensureMembers:   false,
			teamSnap:        nil,
			wantStatus:      StatusFailed,
			wantErrContains: "failed to check existence of channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl, ch, ts := newMocks(t)
			defer ctrl.Finish()
			_ = ts

			cc := &channelCreator{chans: ch}

			body := &createChannelBody{
				TeamRef:    "TeamA",
				ChannelRef: "ChanA",
				MemberRefs: []string{"uBad"},
				OwnerRefs:  []string{},
			}

			ch.EXPECT().
				Get(gomock.Any(), "TeamA", "ChanA").
				DoAndReturn(func(_ context.Context, _ string, _ string) (*models.Channel, error) {
					if tt.existsErr == nil {
						return &models.Channel{ID: "existing"}, nil
					}
					return nil, tt.existsErr
				}).
				Times(1)

			act := cc.planActionForBody(context.Background(), body, tt.ensureMembers, tt.teamSnap)
			require.NotNil(t, act)
			require.NotNil(t, act.result)

			assert.Equal(t, tt.wantStatus, act.result.Status)
			if tt.wantErrContains != "" {
				require.Error(t, act.result.Error)
				assert.Contains(t, act.result.Error.Error(), tt.wantErrContains)
			}
		})
	}
}
