package retriever

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
	"github.com/pzsp-teams/teams-cli/internal/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type sutDeps struct {
	teams    *testutil.MockTeamsService
	channels *testutil.MockChannelsService
}

const team1 = "team-1"
const channelID = "chan-1"

func newSUT(t *testing.T) (Retriever, context.Context, sutDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	d := sutDeps{
		teams:    testutil.NewMockTeamsService(ctrl),
		channels: testutil.NewMockChannelsService(ctrl),
	}
	return NewRetriever(d.teams, d.channels), context.Background(), d
}

func mkMsg(messageID string) *search.SearchResult {
	return &search.SearchResult{
		TeamID:    testutil.Ptr(team1),
		ChannelID: testutil.Ptr(channelID),
		Message: &models.Message{
			ID: messageID,
		},
	}
}

func TestGetMessages_Paginates_ResolvesNames_AndCachesLookups(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	gomock.InOrder(
		d.channels.EXPECT().
			SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				_ context.Context,
				tRef, cRef *string,
				opts *search.SearchMessagesOptions,
				cfg *search.SearchConfig,
			) (*search.SearchResults, error) {
				require.Nil(t, tRef)
				require.Nil(t, cRef)

				require.NotNil(t, cfg)
				require.Equal(t, coreretriever.WorkersCount, cfg.MaxWorkers)

				require.NotNil(t, opts)
				require.True(t, opts.NotFromMe)

				require.NotNil(t, opts.StartTime)
				require.Equal(t, timeRange.Start, *opts.StartTime)
				require.NotNil(t, opts.EndTime)
				require.Equal(t, timeRange.End, *opts.EndTime)

				require.NotNil(t, opts.SearchPage)
				require.NotNil(t, opts.SearchPage.From)
				require.Equal(t, int32(0), *opts.SearchPage.From)
				require.NotNil(t, opts.SearchPage.Size)
				require.Equal(t, int32(25), *opts.SearchPage.Size)

				return &search.SearchResults{
					Messages: []*search.SearchResult{
						mkMsg("m1"),
						mkMsg("m2"),
					},
					NextFrom: testutil.Ptr(int32(25)),
				}, nil
			}),

		d.teams.EXPECT().
			Get(gomock.Any(), "team-1").
			Return(&models.Team{ID: "team-1", DisplayName: "Team One"}, nil).
			Times(1),

		d.channels.EXPECT().
			Get(gomock.Any(), "team-1", "chan-1").
			Return(&models.Channel{ID: "chan-1", Name: "General"}, nil).
			Times(1),

		d.channels.EXPECT().
			SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				_ context.Context,
				_ *string, _ *string,
				opts *search.SearchMessagesOptions,
				_ *search.SearchConfig,
			) (*search.SearchResults, error) {
				require.NotNil(t, opts.SearchPage.From)
				require.Equal(t, int32(25), *opts.SearchPage.From)

				return &search.SearchResults{
					Messages: []*search.SearchResult{
						mkMsg("m3"),
					},
					NextFrom: nil,
				}, nil
			}),
	)

	out, err := r.GetMessages(ctx, timeRange, nil, nil)
	require.NoError(t, err)
	require.Len(t, out, 3)

	require.Equal(t, "Team One", out[0].TeamName)
	require.Equal(t, "team-1", out[0].TeamID)
	require.Equal(t, "General", out[0].ChannelName)
	require.Equal(t, "chan-1", out[0].ChannelID)
	require.Equal(t, "m1", out[0].Message.ID)

	require.Equal(t, "m2", out[1].Message.ID)
	require.Equal(t, "m3", out[2].Message.ID)
}

func TestGetMessages_NormalizesEmptyRefsToNil(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	empty := ""
	d.channels.EXPECT().
		SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{Messages: nil, NextFrom: nil}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, &empty, &empty)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestGetMessages_PassesNonEmptyRefsThrough(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	teamRef := "TeamA"
	channelRef := "General"

	d.channels.EXPECT().
		SearchMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			_ context.Context,
			tRef, cRef *string,
			_ *search.SearchMessagesOptions,
			_ *search.SearchConfig,
		) (*search.SearchResults, error) {
			require.NotNil(t, tRef)
			require.Equal(t, "TeamA", *tRef)
			require.NotNil(t, cRef)
			require.Equal(t, "General", *cRef)
			return &search.SearchResults{Messages: nil, NextFrom: nil}, nil
		}).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, &teamRef, &channelRef)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestGetMessages_PropagatesSearchError(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	exp := errors.New("boom")
	d.channels.EXPECT().
		SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
		Return(nil, exp).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil, nil)
	require.ErrorIs(t, err, exp)
	require.Nil(t, out)
}

func TestGetMessages_SkipsMessagesWithoutTeamOrChannelID(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	msgMissingTeam := &search.SearchResult{
		TeamID:    nil,
		ChannelID: testutil.Ptr("chan-1"),
		Message:   &models.Message{ID: "x"},
	}
	msgMissingChannel := &search.SearchResult{
		TeamID:    testutil.Ptr("team-1"),
		ChannelID: nil,
		Message:   &models.Message{ID: "y"},
	}

	d.channels.EXPECT().
		SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{msgMissingTeam, msgMissingChannel},
			NextFrom: nil,
		}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil, nil)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestGetMessages_IgnoresTeamAndChannelLookupErrors(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	d.channels.EXPECT().
		SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{
				mkMsg("m1"),
			},
			NextFrom: nil,
		}, nil).
		Times(1)

	d.teams.EXPECT().
		Get(gomock.Any(), "team-1").
		Return(nil, errors.New("team lookup failed")).
		Times(1)

	d.channels.EXPECT().
		Get(gomock.Any(), "team-1", "chan-1").
		Return(nil, errors.New("channel lookup failed")).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)

	require.Equal(t, "team-1", out[0].TeamID)
	require.Equal(t, "chan-1", out[0].ChannelID)
	require.Equal(t, "", out[0].TeamName)
	require.Equal(t, "", out[0].ChannelName)
	require.Equal(t, "m1", out[0].Message.ID)
}

func TestGetMessages_BreaksIfNextFromDoesNotAdvance(t *testing.T) {
	t.Parallel()

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	r, ctx, d := newSUT(t)

	d.channels.EXPECT().
		SearchMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{
				mkMsg("m1"),
			},
			NextFrom: testutil.Ptr(int32(0)),
		}, nil).
		Times(1)

	d.teams.EXPECT().
		Get(gomock.Any(), "team-1").
		Return(&models.Team{ID: "team-1", DisplayName: "Team One"}, nil).
		Times(1)

	d.channels.EXPECT().
		Get(gomock.Any(), "team-1", "chan-1").
		Return(&models.Channel{ID: "chan-1", Name: "General"}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "m1", out[0].Message.ID)
}
