package retriever

import (
	"context"
	"fmt"
	"strings"

	f "github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	lib_channels "github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
	lib_teams "github.com/pzsp-teams/lib/teams"
)

type teamChannelMap struct {
	Team     *models.Team
	Channels []*models.Channel
}

// Retriever defines the interface for retrieving channel messages
type Retriever interface {
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*ChannelMessageWithContext, error)
}

// retriever retrieves messages from channels within a time range
type retriever struct {
	teamsService    lib_teams.Service
	channelsService lib_channels.Service
	formatter       f.Formatter
}

// NewRetriever creates a new channel message retriever
func NewRetriever(teamsService lib_teams.Service, channelsService lib_channels.Service, formatter f.Formatter) Retriever {
	return &retriever{
		teamsService:    teamsService,
		channelsService: channelsService,
		formatter:       formatter,
	}
}

func (r *retriever) filterOutArchivedTeams(teams []*models.Team) []*models.Team {
	activeTeams := teams[:0]
	for _, team := range teams {
		if !team.IsArchived {
			activeTeams = append(activeTeams, team)
		}
	}
	return activeTeams
}

func (r *retriever) getActiveTeams(ctx context.Context) ([]*models.Team, error) {
	teams, err := r.teamsService.ListMyJoined(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListingTeamsFailed, err)
	}
	teams = r.filterOutArchivedTeams(teams)
	if len(teams) == 0 {
		return nil, ErrNoTeamsFound
	}
	return teams, nil
}

func (r *retriever) getChannels(ctx context.Context, teams []*models.Team) ([]teamChannelMap, error) {
	var result []teamChannelMap
	for _, team := range teams {
		channels, err := r.channelsService.ListChannels(ctx, team.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrListingChannelsFailed, team.DisplayName, err)
		}
		if len(channels) > 0 {
			result = append(result, teamChannelMap{
				Team:     team,
				Channels: channels,
			})
		}
	}
	if len(result) == 0 {
		return nil, ErrNoChannelsFound
	}
	return result, nil
}

func (r *retriever) getMessagesInTimeRange(ctx context.Context, teamChannelMaps []teamChannelMap, timeRange coreretriever.TimeRange) ([]*ChannelMessageWithContext, error) {
	type channelJob struct {
		Team    *models.Team
		Channel *models.Channel
	}

	var jobs []channelJob
	for _, tcm := range teamChannelMaps {
		for _, channel := range tcm.Channels {
			jobs = append(jobs, channelJob{
				Team:    tcm.Team,
				Channel: channel,
			})
		}
	}

	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}

	results := coreretriever.ExecuteJobs(jobs, coreretriever.WorkersCount, func(job channelJob) ([]*ChannelMessageWithContext, error) {
		messages, err := r.channelsService.ListMessages(ctx, job.Team.ID, job.Channel.ID, opts, false)
		if err != nil && !strings.Contains(err.Error(), "403") {
			return nil, fmt.Errorf("%w: team=%s channel=%s: %v",
				ErrListingMessagesFailed, job.Team.DisplayName, job.Channel.Name, err)
		}

		var filteredMessages []*ChannelMessageWithContext
		for _, message := range messages {
			if message.CreatedDateTime.After(timeRange.Start) && message.CreatedDateTime.Before(timeRange.End) {
				if message.ContentType == models.MessageContentTypeHTML {
					message.Content = r.formatter.Format(message.Content)
				}

				filteredMessages = append(filteredMessages, &ChannelMessageWithContext{
					TeamName:    job.Team.DisplayName,
					TeamID:      job.Team.ID,
					ChannelName: job.Channel.Name,
					ChannelID:   job.Channel.ID,
					Message:     message,
				})
			}
		}

		return filteredMessages, nil
	})

	return coreretriever.AggregateResults(results)
}

// GetMessages retrieves messages from all channels within the specified time range
func (r *retriever) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*ChannelMessageWithContext, error) {
	activeTeams, err := r.getActiveTeams(ctx)
	if err != nil {
		return nil, err
	}
	teamChannelMaps, err := r.getChannels(ctx, activeTeams)
	if err != nil {
		return nil, err
	}
	messages, err := r.getMessagesInTimeRange(ctx, teamChannelMaps, timeRange)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
