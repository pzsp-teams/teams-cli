package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/channels"
)

// SendResult represents the result of sending a message to a channel
type SendResult struct {
	ChannelRef string
	Message    *channels.Message
	Error      error
}

// Sender defines the interface for sending messages to channels
type Sender interface {
	SendToChannels(ctx context.Context, teamRef string, messages map[string]string) []SendResult
}

// ChannelSender wraps the channels service from the library
type ChannelSender struct {
	channelService *channels.Service
}

// NewChannelSender creates a new ChannelSender with the provided channels service
func NewChannelSender(channelService *channels.Service) *ChannelSender {
	return &ChannelSender{
		channelService: channelService,
	}
}

// SendToChannels sends messages to multiple channels within a team
//
// teamRef: team name or ID
// messages: map of channel reference (name or ID) to message content
// Returns a slice of SendResult containing the outcome for each channel
func (s *ChannelSender) SendToChannels(ctx context.Context, teamRef string, messages map[string]string) []SendResult {
	results := make([]SendResult, 0, len(messages))
	logger := initializers.Logger

	logger.Info("Starting bulk message send",
		"team", teamRef,
		"channel_count", len(messages))

	for channelRef, content := range messages {
		result := SendResult{
			ChannelRef: channelRef,
		}

		messageBody := channels.MessageBody{
			Content:     content,
			ContentType: channels.MessageContentTypeHTML,
		}

		logger.Debug("Sending message to channel",
			"team", teamRef,
			"channel", channelRef)

		msg, err := s.channelService.SendMessage(ctx, teamRef, channelRef, messageBody)
		if err != nil {
			logger.Error("Failed to send message",
				"team", teamRef,
				"channel", channelRef,
				"error", err)
			result.Error = fmt.Errorf("failed to send to channel %s: %w", channelRef, err)
		} else {
			logger.Info("Message sent successfully",
				"team", teamRef,
				"channel", channelRef,
				"message_id", msg.ID)
			result.Message = msg
		}

		results = append(results, result)
	}

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}

	logger.Info("Bulk message send completed",
		"team", teamRef,
		"total", len(messages),
		"successful", successCount,
		"failed", len(messages)-successCount)

	return results
}
