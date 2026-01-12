package messages

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/cli/internal/chats/sender"
	"github.com/pzsp-teams/cli/internal/client"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// SendMessages handles sending messages to chats.
func SendMessages(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	template, _ := flags["template"].(string)
	templateData, _ := flags["data"].(string)
	message, _ := flags["message"].(string)
	messageFile, _ := flags["message-file"].(string)

	// Similar handling for slice type as in channels
	var chats []string
	if val, ok := flags["chats"].([]string); ok {
		chats = val
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
		chats,
		internalcommon.ParseTemplateAndData,
	)
	if err != nil {
		return nil, err
	}

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	log.Info("Sending messages to chats", "count", len(processed.Messages), "dryRun", dryRun)
	results := c.Chats.Send(ctx, processed.Messages, dryRun, ignoreErrors)

	printChatResults(w, results, dryRun)

	return results, nil
}

func printChatResults(w io.Writer, results []sender.ChatSendResult, dryRun bool) {
	if dryRun {
		for _, res := range results {
			if res.GetError() != nil {
				_, _ = fmt.Fprintf(w, "Would fail - chat: %s, error: %v\n", res.GetRef(), res.GetError())
			} else {
				_, _ = fmt.Fprintf(w, "Would send - chat: %s, message: %s\n", res.GetRef(), res.GetMessage())
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.GetError() != nil {
				_, _ = fmt.Fprintf(w, "Failed - chat: %s, error: %v\n", res.GetRef(), res.GetError())
			} else {
				successCount++
			}
		}
		_, _ = fmt.Fprintf(w, "Send complete - successful: %d, total: %d\n", successCount, len(results))
	}
}
