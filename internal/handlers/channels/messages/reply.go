package messages

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// ReplyToMessage handles sending a reply to a message.
func ReplyToMessage(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)
	channel, _ := flags["channel"].(string)
	messageID, _ := flags["message-id"].(string)
	message, _ := flags["message"].(string)
	file, _ := flags["message-file"].(string)

	log := initializers.Logger

	content, err := internalcommon.GetMessageContent(message, file)
	if err != nil {
		return nil, err
	}

	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	log.Info("Sending reply to message", "team", team, "channel", channel, "messageID", messageID)
	result := c.Channels.SendReply(ctx, team, channel, messageID, content)

	if result.Error != nil {
		_, _ = fmt.Fprintf(w, "Failed to send reply: %v\n", result.Error)
		return nil, result.Error
	}

	_, _ = fmt.Fprintf(w, "Reply sent successfully to message %s in channel %s\n", messageID, channel)
	return result, nil
}
