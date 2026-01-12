package messages

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/teams-cli/cmd/common"
	"github.com/pzsp-teams/teams-cli/internal/channels/sender"
	"github.com/pzsp-teams/teams-cli/internal/client"
	internalcommon "github.com/pzsp-teams/teams-cli/internal/common"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
)

// SendMessages handles sending messages to channels.
func SendMessages(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	template, _ := flags["template"].(string)
	templateData, _ := flags["data"].(string)
	message, _ := flags["message"].(string)
	messageFile, _ := flags["message-file"].(string)
	team, _ := flags["team"].(string)

	var channels []string
	if val, ok := flags["channels"].([]string); ok {
		channels = val
	}

	dryRun, _ := flags["dry-run"].(bool)
	ignoreErrors, _ := flags["ignore-errors"].(bool)

	log := initializers.Logger

	inputFlags := common.MessageInputFlags{
		Template:     template,
		TemplateData: templateData,
		Message:      message,
		MessageFile:  messageFile,
	}

	processed, err := common.ProcessMessageFlags(
		inputFlags,
		channels,
		internalcommon.ParseTemplateAndData,
	)
	if err != nil {
		return nil, err
	}

	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	log.Info("Sending messages to channels", "team", team, "count", len(processed.Messages), "dryRun", dryRun)
	results := c.Channels.Send(ctx, team, processed.Messages, dryRun, ignoreErrors)

	printChannelResults(w, results, dryRun)

	return results, nil
}

func printChannelResults(w io.Writer, results []sender.ChannelSendResult, dryRun bool) {
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				_, _ = fmt.Fprintf(w, "Would fail - channel: %s, error: %v\n", res.ChannelRef, res.Error)
			} else {
				_, _ = fmt.Fprintf(w, "Would send - channel: %s, message: %s\n", res.ChannelRef, res.Message)
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.Error != nil {
				_, _ = fmt.Fprintf(w, "Failed - channel: %s, error: %v\n", res.ChannelRef, res.Error)
			} else {
				successCount++
			}
		}
		_, _ = fmt.Fprintf(w, "Send complete - successful: %d, total: %d\n", successCount, len(results))
	}
}
