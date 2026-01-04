package retriever

import (
	"context"
	"fmt"
	"strings"

	lib_channels "github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
	lib_teams "github.com/pzsp-teams/lib/teams"
)

type teamChannels = map[string][]string

// Retriever defines the interface for retrieving channel messages
type Retriever interface {
	GetMessages(ctx context.Context, timeRange TimeRange) ([]*DisplayMessageInfo, error)
}

// retriever retrieves messages from channels within a time range
type retriever struct {
	teamsService    lib_teams.Service
	channelsService lib_channels.Service
}

// NewRetriever creates a new channel message retriever
func NewRetriever(teamsService lib_teams.Service, channelsService lib_channels.Service) Retriever {
	return &retriever{
		teamsService:    teamsService,
		channelsService: channelsService,
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

func (r *retriever) getChannels(ctx context.Context, teams []*models.Team) (teamChannels, error) {
	teamChannels := make(teamChannels)
	for _, team := range teams {
		channels, err := r.channelsService.ListChannels(ctx, team.DisplayName)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrListingChannelsFailed, team.DisplayName, err)
		}
		channelNames := make([]string, len(channels))
		for i, channel := range channels {
			channelNames[i] = channel.Name
		}
		if len(channelNames) > 0 {
			teamChannels[team.DisplayName] = channelNames
		}
	}
	if len(teamChannels) == 0 {
		return nil, ErrNoChannelsFound
	}
	return teamChannels, nil
}

func (r *retriever) getMessagesInTimeRange(ctx context.Context, teamChannels teamChannels, timeRange TimeRange) ([]*DisplayMessageInfo, error) {
	var messagesInfo []*DisplayMessageInfo
	top := int32(30)
	opts := &models.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}

	for team, channels := range teamChannels {
		for _, channel := range channels {
			messages, err := r.channelsService.ListMessages(ctx, team, channel, opts)
			if err != nil && !strings.Contains(err.Error(), "403") {
				return nil, fmt.Errorf("%w: team=%s channel=%s: %v",
					ErrListingMessagesFailed, team, channel, err)
			}

			for _, message := range messages {
				if message.CreatedDateTime.After(timeRange.Start) && message.CreatedDateTime.Before(timeRange.End) {
					messagesInfo = append(messagesInfo, &DisplayMessageInfo{
						TeamName:    team,
						ChannelName: channel,
						Message:     message,
					})
				}
			}
		}
	}

	return messagesInfo, nil
}

// GetMessages retrieves messages from all channels within the specified time range
func (r *retriever) GetMessages(ctx context.Context, timeRange TimeRange) ([]*DisplayMessageInfo, error) {
	activeTeams, err := r.getActiveTeams(ctx)
	if err != nil {
		return nil, err
	}
	teamChannels, err := r.getChannels(ctx, activeTeams)
	if err != nil {
		return nil, err
	}
	messagesInfo, err := r.getMessagesInTimeRange(ctx, teamChannels, timeRange)
	if err != nil {
		return nil, err
	}
	return messagesInfo, nil
}
