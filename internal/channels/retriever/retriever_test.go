package retriever

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetMessages_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	channels := []*models.Channel{
		{ID: "channel1-id", Name: "General"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(channels, nil)

	messages := []*models.Message{
		{
			ID:              "msg1",
			Content:         "<p>Hello from channel</p>",
			ContentType:     models.MessageContentTypeHTML,
			CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:              "msg2",
			Content:         "Plain text",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
	}
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "channel1-id", opts, false).
		Return(messages, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// First message should be formatted with context
	assert.Equal(t, "Team1", result[0].TeamName)
	assert.Equal(t, "team1-id", result[0].TeamID)
	assert.Equal(t, "General", result[0].ChannelName)
	assert.Equal(t, "channel1-id", result[0].ChannelID)
	assert.Equal(t, "msg1", result[0].Message.ID)
	assert.Equal(t, "Hello from channel", result[0].Message.Content)

	// Second message should not be formatted (plain text)
	assert.Equal(t, "msg2", result[1].Message.ID)
	assert.Equal(t, "Plain text", result[1].Message.Content)
}

func TestGetMessages_NoTeamsFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return([]*models.Team{}, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNoTeamsFound)
}

func TestGetMessages_ArchivedTeamsFiltered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "archived-team-id", DisplayName: "ArchivedTeam", IsArchived: true},
		{ID: "active-team-id", DisplayName: "ActiveTeam", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	channels := []*models.Channel{
		{ID: "channel1-id", Name: "General"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "active-team-id").
		Return(channels, nil)

	messages := []*models.Message{
		{
			ID:              "msg1",
			Content:         "Active team message",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "active-team-id", "channel1-id", opts, false).
		Return(messages, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "ActiveTeam", result[0].TeamName)
	assert.Equal(t, "Active team message", result[0].Message.Content)
}

func TestGetMessages_NoChannelsFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return([]*models.Channel{}, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNoChannelsFound)
}

func TestGetMessages_TeamsServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	expectedError := errors.New("teams API error")
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(nil, expectedError)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrListingTeamsFailed)
	assert.Contains(t, err.Error(), "teams API error")
}

func TestGetMessages_ChannelsServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	expectedError := errors.New("channels API error")
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(nil, expectedError)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrListingChannelsFailed)
	assert.Contains(t, err.Error(), "channels API error")
}

func TestGetMessages_MessagesServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	channels := []*models.Channel{
		{ID: "channel1-id", Name: "General"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(channels, nil)

	expectedError := errors.New("messages API error")
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "channel1-id", opts, false).
		Return(nil, expectedError)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrListingMessagesFailed)
	assert.Contains(t, err.Error(), "messages API error")
}

func TestGetMessages_403ErrorIgnored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	channels := []*models.Channel{
		{ID: "channel1-id", Name: "General"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(channels, nil)

	forbiddenError := errors.New("403 Forbidden")
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "channel1-id", opts, false).
		Return(nil, forbiddenError)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetMessages_TimeRangeFiltering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	channels := []*models.Channel{
		{ID: "channel1-id", Name: "General"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(channels, nil)

	messages := []*models.Message{
		{
			ID:              "msg1",
			Content:         "Before range",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2023, 12, 31, 23, 0, 0, 0, time.UTC), // Before start
		},
		{
			ID:              "msg2",
			Content:         "In range 1",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), // In range
		},
		{
			ID:              "msg3",
			Content:         "In range 2",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC), // In range
		},
		{
			ID:              "msg4",
			Content:         "After range",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 2, 1, 0, 0, 0, time.UTC), // After end
		},
	}
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "channel1-id", opts, false).
		Return(messages, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 2) // Only messages in range

	assert.Equal(t, "msg2", result[0].Message.ID)
	assert.Equal(t, "In range 1", result[0].Message.Content)

	assert.Equal(t, "msg3", result[1].Message.ID)
	assert.Equal(t, "In range 2", result[1].Message.Content)
}

func TestGetMessages_MultipleTeamsAndChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamsService := testutil.NewMockTeamsService(ctrl)
	mockChannelsService := testutil.NewMockChannelsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	teams := []*models.Team{
		{ID: "team1-id", DisplayName: "Team1", IsArchived: false},
		{ID: "team2-id", DisplayName: "Team2", IsArchived: false},
	}
	mockTeamsService.EXPECT().
		ListMyJoined(ctx).
		Return(teams, nil)

	// Team1 channels
	team1Channels := []*models.Channel{
		{ID: "general-id", Name: "General"},
		{ID: "random-id", Name: "Random"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team1-id").
		Return(team1Channels, nil)

	team2Channels := []*models.Channel{
		{ID: "announcements-id", Name: "Announcements"},
	}
	mockChannelsService.EXPECT().
		ListChannels(ctx, "team2-id").
		Return(team2Channels, nil)

	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}

	team1GeneralMessages := []*models.Message{
		{
			ID:              "t1g1",
			Content:         "Team1 General",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "general-id", opts, false).
		Return(team1GeneralMessages, nil)

	team1RandomMessages := []*models.Message{
		{
			ID:              "t1r1",
			Content:         "Team1 Random",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team1-id", "random-id", opts, false).
		Return(team1RandomMessages, nil)

	team2AnnouncementsMessages := []*models.Message{
		{
			ID:              "t2a1",
			Content:         "Team2 Announcements",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
		},
	}
	mockChannelsService.EXPECT().
		ListMessages(ctx, "team2-id", "announcements-id", opts, false).
		Return(team2AnnouncementsMessages, nil)

	retriever := NewRetriever(mockTeamsService, mockChannelsService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 3) // Messages from all channels

	messageIDs := make([]string, len(result))
	for i, msg := range result {
		messageIDs[i] = msg.Message.ID
	}
	assert.Contains(t, messageIDs, "t1g1")
	assert.Contains(t, messageIDs, "t1r1")
	assert.Contains(t, messageIDs, "t2a1")
}
