package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/channels"
)

// SendResult represents the result of sending a message to a channel
type SendResult struct {
	ChannelRef string
	Message    *channels.Message
	Error      error
}

// SendToChannels sends messages to multiple channels within a team
//
// teamRef: team name or ID
// messages: map of channel reference (name or ID) to message content
// Returns a slice of SendResult containing the outcome for each channel
func SendToChannels(ctx context.Context, c *client.TeamsClient, teamRef string, messages map[string]string) []SendResult {
	results := make([]SendResult, 0, len(messages))
	logger := initializers.Logger

	logger.Info("Starting bulk message send",
		"team", teamRef,
		"channel_count", len(messages))

	for channelRef, content := range messages {
		result := SendResult{
			ChannelRef: channelRef,
		}

		// Wrap message in MessageBody type from library
		messageBody := channels.MessageBody{
			Content:     content,
			ContentType: channels.MessageContentTypeHTML,
		}

		logger.Debug("Sending message to channel",
			"team", teamRef,
			"channel", channelRef)

		// Send message using library's SendMessage method through the embedded client
		msg, err := c.Channels.SendMessage(ctx, teamRef, channelRef, messageBody)
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
