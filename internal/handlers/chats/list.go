package chats

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// ListChats handles listing all chats.
func ListChats(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	log := initializers.Logger

	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	log.Info("Retrieving chats")
	chats, err := c.Chats.List(ctx)
	if err != nil {
		log.Error("Failed to retrieve chats", "error", err)
		return nil, err
	}

	log.Info("Retrieved chats", "count", len(chats))

	if len(chats) == 0 {
		_, _ = fmt.Fprintln(w, "No chats found")
		return nil, nil
	}

	_, _ = fmt.Fprintf(w, "Found %d chats:\n\n", len(chats))
	for i, chat := range chats {
		_, _ = fmt.Fprintf(w, "Chat %d:\n", i+1)
		_, _ = fmt.Fprintf(w, "  ID: %s\n", chat.ID)
		if chat.Topic != nil {
			_, _ = fmt.Fprintf(w, "  Topic: %s\n", *chat.Topic)
		}
		_, _ = fmt.Fprintf(w, "  Type: %s\n", chat.Type)
	}

	return chats, nil
}
