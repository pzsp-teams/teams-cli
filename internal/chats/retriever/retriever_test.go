package retriever

import (
	"context"
	"errors"
	"testing"
	"time"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/cli/internal/utils"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type sutDeps struct {
	chats *testutil.MockChatsService
}

func newSUT(t *testing.T) (Retriever, context.Context, sutDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	d := sutDeps{
		chats: testutil.NewMockChatsService(ctrl),
	}
	return NewRetriever(d.chats), context.Background(), d
}

func mkMsg(chatID string, msgID string, fromDisplayName string) *search.SearchResult {
	var from *models.MessageFrom
	if fromDisplayName != "" {
		from = &models.MessageFrom{DisplayName: fromDisplayName}
	}
	return &search.SearchResult{
		ChatID: testutil.Ptr(chatID),
		Message: &models.Message{
			ID:   msgID,
			From: from,
		},
	}
}

func mkMsgNoChatID(msgID string) *search.SearchResult {
	return &search.SearchResult{
		ChatID: nil,
		Message: &models.Message{
			ID: msgID,
		},
	}
}

func pickOneOnOneChatID(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"user1@contoso.com",
		"someone@example.com",
		"john.doe@pw.edu.pl",
		"user@domain.local",
	}
	for _, id := range candidates {
		if _, ok := utils.GetChatRef(id).(chats.OneOnOneChatRef); ok {
			return id
		}
	}
	ref := utils.GetChatRef(candidates[0])
	t.Fatalf("could not find OneOnOne chat id candidate; utils.GetChatRef(%q) -> %T", candidates[0], ref)
	return ""
}

func pickGroupChatID(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"19:abcdef1234567890@thread.v2",
		"19:groupchat@thread.v2",
		"19:xyz@unq.gbl.spaces",
	}
	for _, id := range candidates {
		if _, ok := utils.GetChatRef(id).(chats.GroupChatRef); ok {
			return id
		}
	}
	ref := utils.GetChatRef(candidates[0])
	t.Fatalf("could not find Group chat id candidate; utils.GetChatRef(%q) -> %T", candidates[0], ref)
	return ""
}

func TestGetMessages_Paginates_GroupChatMetaCached(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	groupChatID := pickGroupChatID(t)
	groupRef := utils.GetChatRef(groupChatID).(chats.GroupChatRef)

	gomock.InOrder(
		d.chats.EXPECT().
			SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				_ context.Context,
				_ chats.ChatRef,
				opts *search.SearchMessagesOptions,
				cfg *search.SearchConfig,
			) (*search.SearchResults, error) {
				require.NotNil(t, cfg)
				require.Equal(t, coreretriever.WorkersCount, cfg.MaxWorkers)

				require.NotNil(t, opts)
				require.True(t, opts.NotFromMe)
				require.Equal(t, timeRange.Start, *opts.StartTime)
				require.Equal(t, timeRange.End, *opts.EndTime)

				require.NotNil(t, opts.SearchPage)
				require.Equal(t, int32(0), *opts.SearchPage.From)
				require.Equal(t, int32(25), *opts.SearchPage.Size)

				return &search.SearchResults{
					Messages: []*search.SearchResult{
						mkMsg(groupChatID, "m1", "Alice"),
						mkMsg(groupChatID, "m2", "Bob"),
					},
					NextFrom: testutil.Ptr(int32(25)),
				}, nil
			}),

		d.chats.EXPECT().
			GetChat(gomock.Any(), groupRef).
			Return(&models.Chat{ID: groupChatID, Topic: testutil.Ptr("Study Group")}, nil).
			Times(1),

		d.chats.EXPECT().
			SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
			DoAndReturn(func(
				_ context.Context,
				_ chats.ChatRef,
				opts *search.SearchMessagesOptions,
				_ *search.SearchConfig,
			) (*search.SearchResults, error) {
				require.NotNil(t, opts.SearchPage.From)
				require.Equal(t, int32(25), *opts.SearchPage.From)

				return &search.SearchResults{
					Messages: []*search.SearchResult{
						mkMsg(groupChatID, "m3", "Eve"),
					},
					NextFrom: nil,
				}, nil
			}),
	)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Len(t, out, 3)

	require.Equal(t, "Study Group", out[0].ChatName)
	require.Equal(t, groupChatID, out[0].ChatID)
	require.Equal(t, "Group", out[0].ChatType)
	require.Equal(t, "m1", out[0].Message.ID)

	require.Equal(t, "m2", out[1].Message.ID)
	require.Equal(t, "m3", out[2].Message.ID)
}

func TestGetMessages_OneOnOne_UsesFromDisplayName(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	oneOnOneID := pickOneOnOneChatID(t)

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{
				mkMsg(oneOnOneID, "m1", "User One"),
			},
			NextFrom: nil,
		}, nil).
		Times(1)

	d.chats.EXPECT().GetChat(gomock.Any(), gomock.Any()).Times(0)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)

	require.Equal(t, "User One", out[0].ChatName)
	require.Equal(t, oneOnOneID, out[0].ChatID)
	require.Equal(t, "OneOnOne", out[0].ChatType)
	require.Equal(t, "m1", out[0].Message.ID)
}

func TestGetMessages_OneOnOne_NoDisplayName_FallsBackToChatID(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	oneOnOneID := pickOneOnOneChatID(t)

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{
				mkMsg(oneOnOneID, "m1", ""),
			},
			NextFrom: nil,
		}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)

	require.Equal(t, oneOnOneID, out[0].ChatName)
	require.Equal(t, oneOnOneID, out[0].ChatID)
	require.Equal(t, "OneOnOne", out[0].ChatType)
}

func TestGetMessages_GroupChat_GetChatError_FallsBackToChatID(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	groupChatID := pickGroupChatID(t)
	groupRef := utils.GetChatRef(groupChatID).(chats.GroupChatRef)

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{mkMsg(groupChatID, "m1", "Alice")},
			NextFrom: nil,
		}, nil).
		Times(1)

	d.chats.EXPECT().
		GetChat(gomock.Any(), groupRef).
		Return(nil, errors.New("boom")).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)

	require.Equal(t, groupChatID, out[0].ChatName)
	require.Equal(t, "Group", out[0].ChatType)
}

func TestGetMessages_GroupChat_NoTopic_FallsBackToChatID(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	groupChatID := pickGroupChatID(t)
	groupRef := utils.GetChatRef(groupChatID).(chats.GroupChatRef)

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{mkMsg(groupChatID, "m1", "Alice")},
			NextFrom: nil,
		}, nil).
		Times(1)

	d.chats.EXPECT().
		GetChat(gomock.Any(), groupRef).
		Return(&models.Chat{ID: groupChatID, Topic: nil}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)

	require.Equal(t, groupChatID, out[0].ChatName)
	require.Equal(t, "Group", out[0].ChatType)
}

func TestGetMessages_SkipsMessagesWithoutChatID(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: []*search.SearchResult{
				mkMsgNoChatID("x"),
				mkMsgNoChatID("y"),
			},
			NextFrom: nil,
		}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestGetMessages_PropagatesSearchError(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	exp := errors.New("search boom")
	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(nil, exp).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.ErrorIs(t, err, exp)
	require.Nil(t, out)
}

func TestGetMessages_BreaksIfNextFromDoesNotAdvance(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), nil, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{
			Messages: nil,
			NextFrom: testutil.Ptr(int32(0)),
		}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, nil)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestGetMessages_PassesChatRefThrough(t *testing.T) {
	t.Parallel()

	r, ctx, d := newSUT(t)

	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatRef := chats.GroupChatRef{Ref: "SomeGroup"}

	d.chats.EXPECT().
		SearchMessages(gomock.Any(), chatRef, gomock.Any(), gomock.Any()).
		Return(&search.SearchResults{Messages: nil, NextFrom: nil}, nil).
		Times(1)

	out, err := r.GetMessages(ctx, timeRange, chatRef)
	require.NoError(t, err)
	require.Empty(t, out)
}
