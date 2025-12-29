package channelcreation

import (
	"context"
	"fmt"
	"strings"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/channels"
)

type channelCreator struct {
	channels channels.Service
}

// ChannelCreator defines the interface for creating channels
type ChannelCreator interface {
	CreateChannels(ctx context.Context, request []map[string]string, ensureMembers bool) []CreateResult
}

// NewChannelCreator creates a new instance of ChannelCreator
func NewChannelCreator(chans channels.Service) *channelCreator {
	return &channelCreator{
		channels: chans,
	}
}

// CreateChannels creates channels based on the provided request data
func (cc *channelCreator) CreateChannels(ctx context.Context, request []map[string]string, ensureMembers bool) []CreateResult {
	results := make([]CreateResult, 0)
	logger := initializers.Logger
	logger.Info("Starting channels creation")
	createChannelBodies := cc.transformRequestToCreateChannelBody(request)
	for _, body := range createChannelBodies {
		exists, err := cc.checkChannelExists(ctx, body.TeamRef, body.ChannelRef)
		if err != nil {
			logger.Error("Failed to check if channel exists", "channel", body.ChannelRef, "team", body.TeamRef, "error", err)
			err = fmt.Errorf("failed to check if channel %s in team %s exists: %w", body.ChannelRef, body.TeamRef, err)
			results = append(results, CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   "",
				Error:       err,
				Status: StatusFailed,
			})
			continue
		}
		if exists {
			logger.Info("Channel already exists, skipping creation", "channel", body.ChannelRef, "team", body.TeamRef)
			if !ensureMembers {
				results = append(results, CreateResult{
					ChannelName: body.ChannelRef,
					ChannelID:   "",
					Error:       nil,
					Status: StatusAlreadyExists,
				})
				continue
			}
			logger.Info("Ensuring members in existing channel", "channel", body.ChannelRef, "team", body.TeamRef)
			err = cc.ensureMembersInChannel(ctx, &body)
			if err != nil {
				logger.Error("Failed to ensure members in existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "error", err)
				err = fmt.Errorf("failed to ensure members in existing channel %s in team %s: %w", body.ChannelRef, body.TeamRef, err)
				results = append(results, CreateResult{
					ChannelName: body.ChannelRef,
					ChannelID:   "",
					Error:       err,
					Status: StatusFailed,
				})
			}
			results = append(results, CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   "",
				Error:       nil,
				Status: StatusMembersEnsured,
				MemberRefs:  body.MemberRefs,
				OwnerRefs:   body.OwnerRefs,
			})
			continue
		}
		channel, err := cc.channels.CreatePrivateChannel(ctx, body.TeamRef, body.ChannelRef, body.MemberRefs, body.OwnerRefs)
		if err != nil {
			logger.Error("Failed to create channel", "channel", body.ChannelRef, "team", body.TeamRef, "error", err)
			err = fmt.Errorf("failed to create channel %s in team %s: %w", body.ChannelRef, body.TeamRef, err)
			results = append(results, CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   "",
				Error:       err,
				Status:      StatusFailed,
			})
		} else {
			logger.Info("Successfully created channel", "channel", body.ChannelRef, "team", body.TeamRef, "channel_id", channel.ID, "members_ref", body.MemberRefs, "owners_ref", body.OwnerRefs)
			results = append(results, CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   channel.ID,
				Error:       nil,
				Status:      StatusCreated,
				MemberRefs:  body.MemberRefs,
				OwnerRefs:   body.OwnerRefs,
			})
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

func (cc *channelCreator) planActions(ctx context.Context, bodies []createChannelBody, ensureMembers bool) []action {
	actions := make([]action, 0)
	for _, body := range bodies {
		exists, err := cc.checkChannelExists(ctx, body.TeamRef, body.ChannelRef)
		if err != nil {
			continue
		}
		if exists {
			if ensureMembers {
				actions = append(actions, action{
					createChannelBody: body,
					run: func(ctx context.Context, body createChannelBody) error {
						return cc.ensureMembersInChannel(ctx, &body)
					},
				})
			}
			continue
		}
		actions = append(actions, action{
			createChannelBody: body,
			run: func(ctx context.Context, body createChannelBody) error {
				_, err := cc.channels.CreatePrivateChannel(ctx, body.TeamRef, body.ChannelRef, body.MemberRefs, body.OwnerRefs)
				return err
			},
		})
	}
	return actions
}

func (cc *channelCreator) transformRequestToCreateChannelBody(data []map[string]string) []createChannelBody {
	out := make([]createChannelBody, 0)
	channelsGroupedByTeamRef := file_readers.GroupBy(data, func(channels map[string]string) string {
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
			out = append(out, createChannelBody{
				TeamRef:    teamRef,
				ChannelRef: channelRef,
				MemberRefs: memberRefs,
				OwnerRefs:  ownerRefs,
			})
		}
	}
	return out
}

func (cc *channelCreator) checkChannelExists(ctx context.Context, teamRef, channelRef string) (bool, error) {
	_, err := cc.channels.Get(ctx, teamRef, channelRef)
	if err != nil {
		if strings.Contains(err.Error(), "[CODE: 404]") || strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (cc *channelCreator) ensureMembersInChannel(ctx context.Context, body *createChannelBody) error {
	logger := initializers.Logger
	for _, memberRef := range body.MemberRefs {
		_, err := cc.channels.AddMember(ctx, body.TeamRef, body.ChannelRef, memberRef, false)
		if err != nil {
			logger.Error("Failed to add member to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "member", memberRef, "error", err)
		} else {
			logger.Info("Successfully added member to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "member", memberRef)
		}
	}
	for _, ownerRef := range body.OwnerRefs {
		_, err := cc.channels.AddMember(ctx, body.TeamRef, body.ChannelRef, ownerRef, true)
		if err != nil {
			logger.Error("Failed to add owner to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "owner", ownerRef, "error", err)
		} else {
			logger.Info("Successfully added owner to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "owner", ownerRef)
		}
	}
	return nil
}
