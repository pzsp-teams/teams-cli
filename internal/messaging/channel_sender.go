package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/templates"
	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
)

type channelSender struct {
	channelService channels.Service
}

// ChannelSender defines the interface for sending messages to channels
type ChannelSender interface {
	SendToChannels(ctx context.Context, teamRef string, messages map[string]string, dryRyn, ignoreError bool) []SendResult
}

// NewChannelSender creates a new ChannelSender with the provided channels service
func NewChannelSender(channelService channels.Service) *channelSender {
	return &channelSender{
		channelService: channelService,
	}
}

// SendToChannels sends messages to multiple channels within a team
//
// teamRef: team name or ID
// messages: map of channel reference (name or ID) to message content
// Returns a slice of SendResult containing the outcome for each channel
func (s *channelSender) SendToChannels(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []SendResult {
	logger := initializers.Logger.With("team", teamRef, "dryRun", dryRun, "ignoreError", ignoreError, "total", len(messages))

	actions := s.planActions(ctx, teamRef, messages)
	var results []SendResult

	logger.Info("Starting bulk message send")

	if dryRun {
		results = s.dryRunActions(actions)
	} else {
		results = s.executeActions(ctx, actions, ignoreError)
	}

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}

	logger.Info("Bulk message send completed",
		"successful", successCount,
		"failed", len(messages)-successCount)

	return results
}

func (s *channelSender) planActions(ctx context.Context, teamRef string, messages map[string]string) []*action {
	actions := make([]*action, 0, len(messages))

	for channelRef, content := range messages {
		processedContent, mentions, err := s.processMentions(ctx, teamRef, channelRef, content)
		if err != nil {
			actions = append(actions, newErrorAction(teamRef, channelRef, content, err))
			continue
		}

		messageData := sendMessageData{
			teamRef:    teamRef,
			channelRef: channelRef,
			body: models.MessageBody{
				Content:     processedContent,
				ContentType: models.MessageContentTypeHTML,
				Mentions:    mentions,
			},
		}

		actions = append(actions, s.newSendAction(&messageData))
	}

	return actions
}

func (s *channelSender) processMentions(ctx context.Context, teamRef, channelRef, content string) (string, []models.Mention, error) {
	rawMentions := templates.ExtractMentions(content)
	if len(rawMentions) == 0 {
		return content, nil, nil
	}

	mentions, err := s.channelService.GetMentions(ctx, teamRef, channelRef, rawMentions)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve mentions for channel %s: %w", channelRef, err)
	}

	return templates.ReplaceMentions(content, mentions), mentions, nil
}

func newErrorAction(teamRef, channelRef, content string, err error) *action {
	return &action{
		sendMessageData: sendMessageData{
			teamRef:    teamRef,
			channelRef: channelRef,
			body: models.MessageBody{
				Content: content,
			},
		},
		run:    nil,
		result: &SendResult{ChannelRef: channelRef, Message: content, Error: err},
	}
}

func (s *channelSender) newSendAction(data *sendMessageData) *action {
	return &action{
		sendMessageData: *data,
		run:             s.sendMessage,
		result:          nil,
	}
}

func (s *channelSender) sendMessage(ctx context.Context, data sendMessageData) *SendResult {
	msg, err := s.channelService.SendMessage(ctx, data.teamRef, data.channelRef, data.body)
	if err != nil {
		return &SendResult{
			ChannelRef: data.channelRef,
			Error:      fmt.Errorf("failed to send to channel %s: %w", data.channelRef, err),
		}
	}
	return &SendResult{
		ChannelRef: data.channelRef,
		Message:    msg.Content,
	}
}

func (s *channelSender) executeActions(ctx context.Context, actions []*action, ignoreError bool) []SendResult {
	logger := initializers.Logger
	results := make([]SendResult, 0, len(actions))
	skipRemaining := false

	for _, action := range actions {
		var result *SendResult

		if skipRemaining {
			logger.Debug("Skipping Message",
				"team", action.teamRef,
				"channel", action.channelRef)

			result = &SendResult{
				ChannelRef: action.channelRef,
				Message:    action.body.Content,
				Error:      ErrMessageSkipped,
			}
		} else if action.run == nil {
			result = action.result
			logger.Error("Failed during planning",
				"team", action.teamRef, "channel", action.channelRef, "error", result.Error)
			if !ignoreError {
				skipRemaining = true
			}
		} else {
			logger.Debug("Sending message to channel",
				"team", action.teamRef,
				"channel", action.channelRef)

			result = action.run(ctx, action.sendMessageData)

			if result.Error != nil {
				logger.Error("Failed to send message",
					"team", action.teamRef, "channel", action.channelRef, "error", result.Error)
				if !ignoreError {
					skipRemaining = true
				}
			} else {
				logger.Info("Message sent successfully", "team", action.teamRef,
					"channel", action.channelRef)
			}
		}

		results = append(results, *result)
	}

	return results
}

func (s *channelSender) dryRunActions(actions []*action) []SendResult {
	results := make([]SendResult, 0, len(actions))
	for _, act := range actions {
		if act.run == nil {
			results = append(results, *act.result)
		} else {
			result := SendResult{ChannelRef: act.channelRef, Message: act.body.Content}
			results = append(results, result)
		}
	}

	return results
}
