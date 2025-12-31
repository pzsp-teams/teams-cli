package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/initializers"
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

	actions := s.planActions(teamRef, messages)
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

func (s *channelSender) planActions(teamRef string, messages map[string]string) []*action {
	actions := make([]*action, 0, len(messages))
	for channelRef, content := range messages {
		result := SendResult{
			ChannelRef: channelRef,
			Message:    content,
		}

		messageBody := models.MessageBody{
			Content:     content,
			ContentType: models.MessageContentTypeHTML,
		}

		messageData := sendMessageData{teamRef, channelRef, messageBody}

		action := action{
			sendMessageData: messageData,
			run: func(ctx context.Context, messageData sendMessageData) *SendResult {
				result = SendResult{}
				msg, err := s.channelService.SendMessage(ctx, messageData.teamRef, messageData.channelRef, messageData.body)
				if err != nil {
					result.Error = fmt.Errorf("failed to send to channel %s: %w", channelRef, err)
				} else {
					result.Message = msg.Content
				}

				return &result
			},
			result: &result,
		}

		actions = append(actions, &action)
	}

	return actions
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
		result := SendResult{ChannelRef: act.channelRef, Message: act.body.Content}
		results = append(results, result)
	}

	return results
}
