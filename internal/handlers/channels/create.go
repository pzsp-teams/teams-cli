package channels

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	channelcreation "github.com/pzsp-teams/cli/internal/channels/creator"
	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// CreateChannels handles creating channels from a file.
func CreateChannels(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)
	data, _ := flags["data"].(string)
	dryRun, _ := flags["dry-run"].(bool)
	ensureInChannels, _ := flags["ensure-in-channels"].(bool)
	ensureInTeam, _ := flags["ensure-in-team"].(bool)

	log := initializers.Logger

	dataFile, err := os.Open(data)
	if err != nil {
		log.Error("Failed to open data file", "file", data, "error", err)
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(data), ".")

	log.Debug("Parsing channels data", "file", data)
	channelData, err := channelcreation.ParseChannelsDataByExtension(dataFile, extension)
	if err != nil {
		log.Error("Failed to parse channels data", "error", err)
		return nil, fmt.Errorf("failed to parse channels data: %w", err)
	}
	_ = dataFile.Close()

	log.Info("Parsed channels data", "channels", len(channelData))

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	log.Info("Creating channels", "count", len(channelData), "dryRun", dryRun)
	results := c.Channels.Create(ctx, team, channelData, ensureInChannels, ensureInTeam, dryRun)

	printChannelCreationResults(w, results, dryRun)

	return results, nil
}

func printChannelCreationResults(w io.Writer, results []channelcreation.CreateResult, dryRun bool) {
	successCount := 0
	for _, res := range results {
		if res.Error != nil {
			fmt.Fprintf(w, "Failed - channel: %s, error: %v\n", res.ChannelName, res.Error)
		} else {
			successCount++
			if dryRun {
				switch res.Status {
				case channelcreation.StatusWouldCreate:
					fmt.Fprintf(w, "[Dry Run] Would create - channel: %s\n", res.ChannelName)
					fmt.Fprintf(w, "Members: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Fprintf(w, "Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusWouldEnsureMembers:
					fmt.Fprintf(w, "[Dry Run] Would ensure members - channel: %s\n", res.ChannelName)
					fmt.Fprintf(w, "Members to ensure: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Fprintf(w, "Owners to ensure: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusAlreadyExists:
					fmt.Fprintf(w, "[Dry Run] Already exists - channel: %s\n", res.ChannelName)
				default:
					fmt.Fprintf(w, "[Dry Run] Processed - channel: %s\n", res.ChannelName)
				}
			} else {
				switch res.Status {
				case channelcreation.StatusCreated:
					fmt.Fprintf(w, "Created - channel: %s\n", res.ChannelName)
					fmt.Fprintf(w, "Members: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Fprintf(w, "Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusAlreadyExists:
					fmt.Fprintf(w, "Already exists - channel: %s\n", res.ChannelName)
				case channelcreation.StatusMembersEnsured:
					fmt.Fprintf(w, "Members ensured - channel: %s\n", res.ChannelName)
					fmt.Fprintf(w, "Members ensured: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Fprintf(w, "Owners ensured: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusPartiallyEnsured:
					fmt.Fprintf(w, "Partially ensured - channel: %s\n", res.ChannelName)
					fmt.Fprintf(w, "Members ensured: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Fprintf(w, "Owners ensured: %s\n", strings.Join(res.OwnerRefs, ", "))
				default:
					fmt.Fprintf(w, "Processed - channel: %s\n", res.ChannelName)
				}
			}
		}
	}

	if dryRun {
		fmt.Fprintf(w, "Dry run completed - successful: %d, total: %d\n", successCount, len(results))
	} else {
		fmt.Fprintf(w, "Channel creation completed - successful: %d, total: %d\n", successCount, len(results))
	}
}
