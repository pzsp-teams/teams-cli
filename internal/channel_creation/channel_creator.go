// TODO: Add ensure members to team when creating channels

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
	CreateChannels(ctx context.Context, request []map[string]string, ensureMembers, dryRun bool) []CreateResult
}

// NewChannelCreator creates a new instance of ChannelCreator
func NewChannelCreator(chans channels.Service) *channelCreator {
	return &channelCreator{
		channels: chans,
	}
}

// CreateChannels creates channels based on the provided request data
func (cc *channelCreator) CreateChannels(ctx context.Context, request []map[string]string, ensureMembers, dryRun bool) []CreateResult {
	logger := initializers.Logger
	logger.Info("Starting channels creation")
	createChannelBodies := cc.transformRequestToCreateChannelBody(request)
	actions := cc.planActions(ctx, createChannelBodies, ensureMembers)
	var results []CreateResult
	if dryRun {
		logger.Info("Dry run enabled - no channels will be created")
		results = cc.dryRunActions(actions)
	} else {
		results = cc.executeActions(ctx, actions)
	}
	logger.Info("Channels creation process completed")
	return results
}

func (cc *channelCreator) executeActions(ctx context.Context, actions []*action) []CreateResult {
	logger := initializers.Logger
	results := make([]CreateResult, 0, len(actions))
	for _, act := range actions {
		result := act.run(ctx, act.createChannelBody)
		if result == nil {
			logger.Error("Action returned nil result", "channel", act.ChannelRef, "team", act.TeamRef)
			result = &CreateResult{
				ChannelName: act.ChannelRef,
				ChannelID:   "",
				Error:       fmt.Errorf("action returned nil result"),
				Status:      StatusFailed,
			}
		}
		logExecutionResult(result, act.TeamRef)
		results = append(results, *result)
	}
	return results
}

func (cc *channelCreator) dryRunActions(actions []*action) []CreateResult {
	results := make([]CreateResult, 0, len(actions))
	for _, act := range actions {
		res := act.result
		if res == nil {
			initializers.Logger.Error("Action has nil result in dry run", "channel", act.ChannelRef, "team", act.TeamRef)
			res = &CreateResult{
				ChannelName: act.ChannelRef,
				ChannelID:   "",
				Error:       fmt.Errorf("action has nil result in dry run"),
				Status:      StatusFailed,
			}
		}
		results = append(results, *res)
		logDryRunResult(res, act.TeamRef)
	}
	return results
}

func (cc *channelCreator) planActions(ctx context.Context, bodies []createChannelBody, ensureMembers bool) []*action {
	actions := make([]*action, 0, len(bodies))
	for i := range bodies {
		act := cc.planActionForBody(ctx, &bodies[i], ensureMembers)
		actions = append(actions, act)
	}
	return actions
}

func (cc *channelCreator) planActionForBody(ctx context.Context, body *createChannelBody, ensureMembers bool) *action {
	exists, err := cc.checkChannelExists(ctx, body.TeamRef, body.ChannelRef)
	if err != nil {
		errToShow := fmt.Errorf("failed to check existence of channel %s in team %s: %w", body.ChannelRef, body.TeamRef, err)
		return staticAction(body, failedResult(errToShow, body))
	}
	if exists {
		if ensureMembers {
			return cc.ensureMembersAction(body)
		}
		return staticAction(body, alreadyExistsResult(body))
	}
	return cc.createChannelAction(body)
}

func (cc *channelCreator) createChannelAction(body *createChannelBody) *action {
	return &action{
		createChannelBody: *body,
		run: func(ctx context.Context, body createChannelBody) *CreateResult {
			channel, err := cc.channels.CreatePrivateChannel(ctx, body.TeamRef, body.ChannelRef, body.MemberRefs, body.OwnerRefs)
			if err != nil {
				err = fmt.Errorf("failed to create channel %s in team %s: %w", body.ChannelRef, body.TeamRef, err)
				return &CreateResult{
					ChannelName: body.ChannelRef,
					ChannelID:   "",
					Error:       err,
					Status:      StatusFailed,
				}
			}
			return &CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   channel.ID,
				Error:       nil,
				Status:      StatusCreated,
				MemberRefs:  body.MemberRefs,
				OwnerRefs:   body.OwnerRefs,
			}
		},
		result: &CreateResult{
			ChannelName: body.ChannelRef,
			ChannelID:   "",
			Error:       nil,
			Status:      StatusWouldCreate,
			MemberRefs:  body.MemberRefs,
			OwnerRefs:   body.OwnerRefs,
		},
	}
}

func (cc *channelCreator) ensureMembersAction(body *createChannelBody) *action {
	return &action{
		createChannelBody: *body,
		run: func(ctx context.Context, body createChannelBody) *CreateResult {
			ensureMembersResult := cc.ensureMembersInChannel(ctx, &body)
			if len(ensureMembersResult.MembersRefsFailed) > 0 || len(ensureMembersResult.OwnerRefsFailed) > 0 {
				return &CreateResult{
					ChannelName: body.ChannelRef,
					ChannelID:   "",
					Error:       errMembersPartiallyEnsured,
					Status:      StatusPartiallyEnsured,
					MemberRefs:  ensureMembersResult.MembersRefsEnsured,
					OwnerRefs:   ensureMembersResult.OwnerRefsEnsured,
				}
			}
			return &CreateResult{
				ChannelName: body.ChannelRef,
				ChannelID:   "",
				Error:       nil,
				Status:      StatusMembersEnsured,
				MemberRefs:  ensureMembersResult.MembersRefsEnsured,
				OwnerRefs:   ensureMembersResult.OwnerRefsEnsured,
			}
		},
		result: &CreateResult{
			ChannelName: body.ChannelRef,
			ChannelID:   "",
			Error:       nil,
			Status:      StatusWouldEnsureMembers,
			MemberRefs:  body.MemberRefs,
			OwnerRefs:   body.OwnerRefs,
		},
	}
}

func logExecutionResult(result *CreateResult, teamRef string) {
	logger := initializers.Logger
	switch result.Status {
	case StatusCreated:
		logger.Info("Channel created successfully", "channel", result.ChannelName, "channel_id", result.ChannelID, "team", teamRef, "status", result.Status, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case StatusAlreadyExists:
		logger.Info("Channel already exists", "channel", result.ChannelName, "team", teamRef, "status", result.Status)
	case StatusMembersEnsured:
		logger.Info("Members ensured in existing channel", "channel", result.ChannelName, "team", teamRef, "status", result.Status, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case StatusPartiallyEnsured:
		logger.Warn("Members partially ensured in existing channel", "channel", result.ChannelName, "team", teamRef, "status", result.Status, "members_refs_ensured", result.MemberRefs, "owner_refs_ensured", result.OwnerRefs)
	case StatusFailed:
		logger.Error("Channel operation failed", "channel", result.ChannelName, "team", teamRef, "error", result.Error, "status", result.Status)
	}
}

func logDryRunResult(result *CreateResult, teamRef string) {
	logger := initializers.Logger
	switch result.Status {
	case StatusWouldCreate:
		logger.Info("Dry run: Channel would be created", "channel", result.ChannelName, "team", teamRef, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case StatusWouldEnsureMembers:
		logger.Info("Dry run: Members would be ensured in existing channel", "channel", result.ChannelName, "team", teamRef, "members_refs", result.MemberRefs, "owner_refs", result.OwnerRefs)
	case StatusAlreadyExists:
		logger.Info("Dry run: Channel already exists", "channel", result.ChannelName, "team", teamRef)
	case StatusFailed:
		logger.Error("Dry run: Channel creation would fail", "channel", result.ChannelName, "team", teamRef, "error", result.Error)
	}
}

func failedResult(err error, body *createChannelBody) *CreateResult {
	return &CreateResult{
		ChannelName: body.ChannelRef,
		ChannelID:   "",
		Error:       err,
		Status:      StatusFailed,
	}
}

func alreadyExistsResult(body *createChannelBody) *CreateResult {
	return &CreateResult{
		ChannelName: body.ChannelRef,
		ChannelID:   "",
		Error:       nil,
		Status:      StatusAlreadyExists,
	}
}

func staticAction(body *createChannelBody, result *CreateResult) *action {
	return &action{
		createChannelBody: *body,
		result:            result,
		run: func(ctx context.Context, body createChannelBody) *CreateResult {
			return result
		},
	}
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

func (cc *channelCreator) ensureMembersInChannel(ctx context.Context, body *createChannelBody) *ensureMembersResult {
	logger := initializers.Logger
	var out ensureMembersResult
	for _, memberRef := range body.MemberRefs {
		_, err := cc.channels.AddMember(ctx, body.TeamRef, body.ChannelRef, memberRef, false)
		if err != nil {
			logger.Error("Failed to add member to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "member", memberRef, "error", err)
			out.MembersRefsFailed = append(out.MembersRefsFailed, memberRef)
		} else {
			logger.Info("Successfully added member to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "member", memberRef)
			out.MembersRefsEnsured = append(out.MembersRefsEnsured, memberRef)
		}
	}
	for _, ownerRef := range body.OwnerRefs {
		_, err := cc.channels.AddMember(ctx, body.TeamRef, body.ChannelRef, ownerRef, true)
		if err != nil {
			logger.Error("Failed to add owner to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "owner", ownerRef, "error", err)
			out.OwnerRefsFailed = append(out.OwnerRefsFailed, ownerRef)
		} else {
			logger.Info("Successfully added owner to existing channel", "channel", body.ChannelRef, "team", body.TeamRef, "owner", ownerRef)
			out.OwnerRefsEnsured = append(out.OwnerRefsEnsured, ownerRef)
		}
	}
	return &out
}
