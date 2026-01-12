package retriever

import (
	"context"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	lib_channels "github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/search"
	lib_teams "github.com/pzsp-teams/lib/teams"
)

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

func (r *retriever) getMessagesInTimeRange(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*ChannelMessageWithContext, error) {
	var aggregatedSearchResults []*search.SearchResults
	var from int32 = 0
	var size int32 = 25
	var tRef, cRef *string
	if teamRef != nil && *teamRef != "" {
		tRef = teamRef
	}
	if channelRef != nil && *channelRef != "" {
		cRef = channelRef
	}
	searchConfig := &search.SearchConfig{
		MaxWorkers: 32,
	}
	for {
		searchOpts := &search.SearchMessagesOptions{
			StartTime: &timeRange.Start,
			EndTime:   &timeRange.End,
			NotFromMe: true,
			SearchPage: &search.SearchPage{
				From: &from,
				Size: &size,
			},
		}

		searchResult, err := r.channelsService.SearchMessages(ctx, tRef, cRef, searchOpts, searchConfig)
		if err != nil {
			return nil, err
		}
		aggregatedSearchResults = append(aggregatedSearchResults, searchResult)
		if searchResult.NextFrom == nil {
			break
		}
		from = *searchResult.NextFrom
	}
	return r.processChannelMessages(aggregatedSearchResults), nil
}

func (r *retriever) processChannelMessages(searchResults []*search.SearchResults) []*ChannelMessageWithContext {
	var results []*ChannelMessageWithContext
	var teamNameByID = make(map[string]string)
	var channelNameByID = make(map[string]string)
	for _, result := range searchResults {
		for _, msg := range result.Messages {
			if msg.TeamID == nil || msg.ChannelID == nil {
				continue
			}
			var teamName, channelName string
			if name, ok := teamNameByID[*msg.TeamID]; ok {
				teamName = name
			} else if msg.TeamID != nil {
				team, err := r.teamsService.Get(context.Background(), *msg.TeamID)
				if err == nil {
					teamName = team.DisplayName
					teamNameByID[*msg.TeamID] = team.DisplayName
				}
			}
			if name, ok := channelNameByID[*msg.ChannelID]; ok {
				channelName = name
			} else if msg.ChannelID != nil {
				channel, err := r.channelsService.Get(context.Background(), *msg.TeamID, *msg.ChannelID)
				if err == nil {
					channelName = channel.Name
					channelNameByID[*msg.ChannelID] = channel.Name
				}
			}
			results = append(results, &ChannelMessageWithContext{
				TeamName:    teamName,
				TeamID:      *msg.TeamID,
				ChannelName: channelName,
				ChannelID:   *msg.ChannelID,
				Message:     msg.Message,
			})
		}
	}
	return results
}

// GetMessages retrieves messages from all channels within the specified time range
func (r *retriever) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*ChannelMessageWithContext, error) {
	messages, err := r.getMessagesInTimeRange(ctx, timeRange, teamRef, channelRef)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
