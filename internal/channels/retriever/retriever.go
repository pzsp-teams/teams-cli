package retriever

import (
	"context"

	lib_channels "github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/search"
	lib_teams "github.com/pzsp-teams/lib/teams"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
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

// GetMessages retrieves messages from channels within the specified time range and optional team/channel filters
func (r *retriever) GetMessages(
	ctx context.Context,
	timeRange coreretriever.TimeRange,
	teamRef, channelRef *string,
) ([]*ChannelMessageWithContext, error) {
	tRef := normalizeRef(teamRef)
	cRef := normalizeRef(channelRef)

	searchConfig := &search.SearchConfig{
		MaxWorkers: coreretriever.WorkersCount,
	}

	teamNameByID := make(map[string]string)
	channelNameByKey := make(map[string]string)

	var results []*ChannelMessageWithContext

	from := int32(0)
	size := int32(25)

	for range 10_000 {
		f := from
		s := size

		searchOpts := &search.SearchMessagesOptions{
			StartTime: &timeRange.Start,
			EndTime:   &timeRange.End,
			NotFromMe: true,
			SearchPage: &search.SearchPage{
				From: &f,
				Size: &s,
			},
		}

		page, err := r.channelsService.SearchMessages(ctx, tRef, cRef, searchOpts, searchConfig)
		if err != nil {
			return nil, err
		}

		r.processPage(ctx, page, teamNameByID, channelNameByKey, &results)

		if page.NextFrom == nil || *page.NextFrom == from {
			break
		}
		from = *page.NextFrom
	}

	return results, nil
}

func normalizeRef(ref *string) *string {
	if ref == nil || *ref == "" {
		return nil
	}
	return ref
}

func (r *retriever) processPage(
	ctx context.Context,
	page *search.SearchResults,
	teamNameByID map[string]string,
	channelNameByKey map[string]string,
	out *[]*ChannelMessageWithContext,
) {
	for _, msg := range page.Messages {
		if msg.TeamID == nil || msg.ChannelID == nil {
			continue
		}

		teamID := *msg.TeamID
		channelID := *msg.ChannelID
		channelKey := teamID + ":" + channelID

		teamName := teamNameByID[teamID]
		if teamName == "" {
			if team, err := r.teamsService.Get(ctx, teamID); err == nil && team != nil {
				teamName = team.DisplayName
				teamNameByID[teamID] = teamName
			}
		}

		channelName := channelNameByKey[channelKey]
		if channelName == "" {
			if ch, err := r.channelsService.Get(ctx, teamID, channelID); err == nil && ch != nil {
				channelName = ch.Name
				channelNameByKey[channelKey] = channelName
			}
		}

		*out = append(*out, &ChannelMessageWithContext{
			TeamName:    teamName,
			TeamID:      teamID,
			ChannelName: channelName,
			ChannelID:   channelID,
			Message:     msg.Message,
		})
	}
}
