package channelcreation

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/channels"
)

type CreateResult struct {
	ChannelName string
	ChannelID   string
	Error 	 	error
}
type channelCreator struct {
	channels channels.Service
}

type ChannelCreator interface {
	CreateChannels(ctx context.Context, request []map[string]string) []CreateResult
}

func NewChannelCreator(channels channels.Service) *channelCreator {
	return &channelCreator{
		channels: channels,
	}
}

func (cc *channelCreator) CreateChannels(ctx context.Context, request []map[string]string) []CreateResult {
	results := make([]CreateResult, 0)
	logger := initializers.Logger
	logger.Info("Starting channels creation")
	channelsGroupedByTeamRef := file_readers.GroupBy(request, func(channels map[string]string) string {
		return channels["team_ref"]
	})
	for teamRef, channelData := range channelsGroupedByTeamRef {
		usersGroupedByChannelRef := file_readers.GroupBy(channelData, func(channel map[string]string) string {
			return channel["channel_ref"]
		})
		for channelRef, usersData := range usersGroupedByChannelRef {
			usersGroupedByRole := file_readers.GroupBy(usersData, func(user map[string]string) string {
				return user["role"]
			})
			memberRefs := make([]string, 0)
			ownerRefs := make([]string, 0)
			for role, users := range usersGroupedByRole {
				for _, user := range users {
					if role == "owner" {
						ownerRefs = append(ownerRefs, user["user_ref"])
					} else {
						memberRefs = append(memberRefs, user["user_ref"])
					}
				}
			}
			channel, err := cc.channels.CreatePrivateChannel(ctx, teamRef, channelRef, memberRefs, ownerRefs)
			if err != nil {
				logger.Error("Failed to create channel", "channel", channelRef, "team", teamRef, "error", err)
				err = fmt.Errorf("failed to create channel %s in team %s: %w", channelRef, teamRef, err)
				results = append(results, CreateResult{
					ChannelName: channelRef,
					ChannelID:   "",
					Error:       err,
				})
			} else {
					logger.Info("Successfully created channel", "channel", channelRef, "team", teamRef, "channel_id", channel.ID)
				results = append(results, CreateResult{
					ChannelName: channelRef,
					ChannelID:   channel.ID,
					Error:       nil,
				})
			}	
		}
	}
	successCount := 0
	for _, result := range results {
		if result.Error == nil {
			successCount++
		}
	}
	logger.Info("Channels creation process completed", "success_count", successCount, "total", len(results))
	return results
}

