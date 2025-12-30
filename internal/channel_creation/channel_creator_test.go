package channelcreation

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/logger"
	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopLogger struct{}

func (noopLogger) With(args ...any) logger.Logger {
	return noopLogger{}
}

func (noopLogger) Debug(msg string, args ...any) {}
func (noopLogger) Info(msg string, args ...any)  {}
func (noopLogger) Warn(msg string, args ...any)  {}
func (noopLogger) Error(msg string, args ...any) {}

var _ logger.Logger = noopLogger{}

func TestMain(m *testing.M) {
	initializers.Logger = noopLogger{}
	m.Run()
}

func row(teamRef, channelRef, role, userRef string) map[string]string {
	return map[string]string{
		"team_ref":    teamRef,
		"channel_ref": channelRef,
		"role":        role,
		"user_ref":    userRef,
	}
}

type channelsServiceStub struct {
	channels.Service

	getFn    func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error)
	createFn func(ctx context.Context, teamRef, channelRef string, memberRefs, ownerRefs []string) (*models.Channel, error)

	getCalls    int
	createCalls int
}

func (s *channelsServiceStub) Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
	s.getCalls++
	if s.getFn == nil {
		return nil, errors.New("Get not configured in stub")
	}
	return s.getFn(ctx, teamRef, channelRef)
}

func (s *channelsServiceStub) CreatePrivateChannel(
	ctx context.Context,
	teamRef, channelRef string,
	memberRefs, ownerRefs []string,
) (*models.Channel, error) {
	s.createCalls++
	if s.createFn == nil {
		return nil, errors.New("CreatePrivateChannel not configured in stub")
	}
	return s.createFn(ctx, teamRef, channelRef, memberRefs, ownerRefs)
}

func TestTransformRequestToCreateChannelBody_GroupsAndRoles(t *testing.T) {
	cc := &channelCreator{}

	in := []map[string]string{
		row("t1", "c1", "member", "u1"),
		row("t1", "c1", "owner", "u2"),
		row("t1", "c1", "member", "u3"),
		row("t1", "c2", "member", "u4"),
		row("t2", "cx", "owner", "u9"),
	}

	got := cc.transformRequestToCreateChannelBody(in)
	require.Len(t, got, 3)

	byKey := map[string]createChannelBody{}
	for _, b := range got {
		byKey[b.TeamRef+"::"+b.ChannelRef] = b
	}

	b := byKey["t1::c1"]
	assert.Equal(t, "t1", b.TeamRef)
	assert.Equal(t, "c1", b.ChannelRef)
	assert.ElementsMatch(t, []string{"u1", "u3"}, b.MemberRefs)
	assert.ElementsMatch(t, []string{"u2"}, b.OwnerRefs)

	b = byKey["t1::c2"]
	assert.ElementsMatch(t, []string{"u4"}, b.MemberRefs)
	assert.Empty(t, b.OwnerRefs)

	b = byKey["t2::cx"]
	assert.Empty(t, b.MemberRefs)
	assert.ElementsMatch(t, []string{"u9"}, b.OwnerRefs)
}

func TestCheckChannelExists_NotFound_ReturnsFalseNil(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return nil, errors.New("[CODE: 404] not found")
		},
	}
	cc := NewChannelCreator(stub)

	exists, err := cc.checkChannelExists(context.Background(), "t", "c")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCheckChannelExists_OtherError_ReturnsError(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return nil, errors.New("boom")
		},
	}
	cc := NewChannelCreator(stub)

	exists, err := cc.checkChannelExists(context.Background(), "t", "c")
	require.Error(t, err)
	assert.False(t, exists)
}

func TestCreateChannels_DryRun_MissingChannel_ReturnsWouldCreate_AndDoesNotCreate(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return nil, errors.New("[CODE: 404] not found")
		},
		createFn: func(ctx context.Context, teamRef, channelRef string, memberRefs, ownerRefs []string) (*models.Channel, error) {
			t.Fatalf("CreatePrivateChannel must not be called in dry-run")
			return nil, nil
		},
	}

	cc := NewChannelCreator(stub)
	req := []map[string]string{
		row("teamA", "chanA", "member", "u1"),
		row("teamA", "chanA", "owner", "u2"),
	}

	res := cc.CreateChannels(context.Background(), req, false, true)
	require.Len(t, res, 1)

	assert.Equal(t, "chanA", res[0].ChannelName)
	assert.Equal(t, StatusWouldCreate, res[0].Status)
	assert.ElementsMatch(t, []string{"u1"}, res[0].MemberRefs)
	assert.ElementsMatch(t, []string{"u2"}, res[0].OwnerRefs)

	assert.Equal(t, 1, stub.getCalls)
	assert.Equal(t, 0, stub.createCalls)
}

func TestCreateChannels_Execute_MissingChannel_CreatesAndReturnsCreated(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return nil, errors.New("[CODE: 404] not found")
		},
		createFn: func(ctx context.Context, teamRef, channelRef string, memberRefs, ownerRefs []string) (*models.Channel, error) {
			assert.Equal(t, "teamA", teamRef)
			assert.Equal(t, "chanA", channelRef)
			assert.ElementsMatch(t, []string{"u1"}, memberRefs)
			assert.ElementsMatch(t, []string{"u2"}, ownerRefs)

			return &models.Channel{ID: "channel-id-123"}, nil
		},
	}

	cc := NewChannelCreator(stub)
	req := []map[string]string{
		row("teamA", "chanA", "member", "u1"),
		row("teamA", "chanA", "owner", "u2"),
	}

	res := cc.CreateChannels(context.Background(), req, false, false)
	require.Len(t, res, 1)

	assert.Equal(t, StatusCreated, res[0].Status)
	assert.Equal(t, "chanA", res[0].ChannelName)
	assert.Equal(t, "channel-id-123", res[0].ChannelID)
	assert.ElementsMatch(t, []string{"u1"}, res[0].MemberRefs)
	assert.ElementsMatch(t, []string{"u2"}, res[0].OwnerRefs)

	assert.Equal(t, 1, stub.getCalls)
	assert.Equal(t, 1, stub.createCalls)
}

func TestCreateChannels_DryRun_ChannelExists_EnsureMembersFalse_ReturnsAlreadyExists(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return &models.Channel{ID: "existing"}, nil
		},
		createFn: func(ctx context.Context, teamRef, channelRef string, memberRefs, ownerRefs []string) (*models.Channel, error) {
			t.Fatalf("CreatePrivateChannel must not be called when channel exists")
			return nil, nil
		},
	}
	cc := NewChannelCreator(stub)

	req := []map[string]string{
		row("teamA", "chanA", "member", "u1"),
	}

	res := cc.CreateChannels(context.Background(), req, false, true)
	require.Len(t, res, 1)

	assert.Equal(t, StatusAlreadyExists, res[0].Status)
	assert.Equal(t, "chanA", res[0].ChannelName)
	assert.Equal(t, 1, stub.getCalls)
	assert.Equal(t, 0, stub.createCalls)
}

func TestCreateChannels_DryRun_ChannelExists_EnsureMembersTrue_ReturnsWouldEnsureMembers(t *testing.T) {
	stub := &channelsServiceStub{
		getFn: func(ctx context.Context, teamRef, channelRef string) (*models.Channel, error) {
			return &models.Channel{ID: "existing"}, nil
		},
	}
	cc := NewChannelCreator(stub)

	req := []map[string]string{
		row("teamA", "chanA", "member", "u1"),
		row("teamA", "chanA", "owner", "u2"),
	}

	res := cc.CreateChannels(context.Background(), req, true, true)
	require.Len(t, res, 1)

	assert.Equal(t, StatusWouldEnsureMembers, res[0].Status)
	assert.Equal(t, "chanA", res[0].ChannelName)
	assert.ElementsMatch(t, []string{"u1"}, res[0].MemberRefs)
	assert.ElementsMatch(t, []string{"u2"}, res[0].OwnerRefs)
}

func TestDryRunActions_WhenActionHasNilResult_ReturnsFailed(t *testing.T) {
	cc := &channelCreator{}

	acts := []*action{
		{
			createChannelBody: createChannelBody{
				TeamRef:    "t",
				ChannelRef: "c",
			},
			result: nil,
		},
	}

	out := cc.dryRunActions(acts)
	require.Len(t, out, 1)

	assert.Equal(t, "c", out[0].ChannelName)
	assert.Equal(t, StatusFailed, out[0].Status)
	require.Error(t, out[0].Error)
}

func TestExecuteActions_WhenRunReturnsNil_ResultIsFailed(t *testing.T) {
	cc := &channelCreator{}

	acts := []*action{
		{
			createChannelBody: createChannelBody{
				TeamRef:    "t",
				ChannelRef: "c",
			},
			run: func(ctx context.Context, body createChannelBody) *CreateResult {
				return nil
			},
		},
	}

	out := cc.executeActions(context.Background(), acts)
	require.Len(t, out, 1)

	assert.Equal(t, "c", out[0].ChannelName)
	assert.Equal(t, StatusFailed, out[0].Status)
	require.Error(t, out[0].Error)
	assert.Contains(t, fmt.Sprint(out[0].Error), "nil result")
}
