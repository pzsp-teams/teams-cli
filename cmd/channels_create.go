package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	channelcreation "github.com/pzsp-teams/cli/internal/channels/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var createChannelsCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Teams channels from a data file",
	Long: `Create multiple Teams channels with members from a data file (YAML/JSON/TOML/CSV).

The data file should contain channel definitions with team_ref, channel_ref, role, and user_ref.

Examples:
  # Create channels from YAML file
  cli channels create --team myteam --data channels.yaml

  # Create channels from JSON file
  cli channels create --team myteam --data channels.json

  # Dry run to preview
  cli channels create --team myteam --data channels.yaml --dry-run
  
  # Ensure members are added to channels if they already exist
  cli channels create --team myteam --data channels.yaml --ensure-in-channels

  # Ensure members are memebers of the team
  cli channels create --team myteam --data channels.yaml --ensure-in-team
  `,
	RunE: runCreateChannels,
}

var (
	teamRef              string
	createChannelsData   string
	createChannelsDryRun bool
	ensureInChannels     bool
	ensureInTeam         bool
)

func init() {
	createChannelsCmd.Flags().StringVar(&teamRef, "team", "", "Name of the team in which to create channels")
	createChannelsCmd.Flags().StringVar(&createChannelsData, "data", "", "Path to channels data file (YAML/JSON/TOML/CSV)")
	createChannelsCmd.Flags().BoolVar(&createChannelsDryRun, "dry-run", false, "Preview without creating channels")
	createChannelsCmd.Flags().BoolVar(&ensureInChannels, "ensure-in-channels", false, "Ensure members are added to channels if they already exist")
	createChannelsCmd.Flags().BoolVar(&ensureInTeam, "ensure-in-team", false, "Ensure members are members of the team")

	if err := createChannelsCmd.MarkFlagRequired("data"); err != nil {
		panic(fmt.Sprintf("failed to mark data flag as required: %v", err))
	}
	if err := createChannelsCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runCreateChannels(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()

	dataFile, err := os.Open(createChannelsData)
	if err != nil {
		log.Error("Failed to open data file", "file", createChannelsData, "error", err)
		return fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(createChannelsData), ".")

	log.Debug("Parsing channels data", "file", createChannelsData)
	channelData, err := channelcreation.ParseChannelsDataByExtension(dataFile, extension)
	if err != nil {
		log.Error("Failed to parse channels data", "error", err)
		return fmt.Errorf("failed to parse channels data: %w", err)
	}
	_ = dataFile.Close()

	log.Info("Parsed channels data", "channels", len(channelData))

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Creating channels", "count", len(channelData), "dryRun", createChannelsDryRun)
	results := teamsClient.Channels.Create(ctx, teamRef, channelData, ensureInChannels, ensureInTeam, createChannelsDryRun)

	printChannelCreationResults(results, createChannelsDryRun)

	return nil
}

func printChannelCreationResults(results []channelcreation.CreateResult, dryRun bool) {
	successCount := 0
	for _, res := range results {
		if res.Error != nil {
			fmt.Printf("Failed - channel: %s, error: %v\n", res.ChannelName, res.Error)
		} else {
			successCount++
			if dryRun {
				switch res.Status {
				case channelcreation.StatusWouldCreate:
					fmt.Printf("[Dry Run] Would create - channel: %s\n", res.ChannelName)
					fmt.Printf("Members: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Printf("Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusWouldEnsureMembers:
					fmt.Printf("[Dry Run] Would ensure members - channel: %s\n", res.ChannelName)
					fmt.Printf("Members to ensure: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Printf("Owners to ensure: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusAlreadyExists:
					fmt.Printf("[Dry Run] Already exists - channel: %s\n", res.ChannelName)
				default:
					fmt.Printf("[Dry Run] Processed - channel: %s\n", res.ChannelName)
				}
			} else {
				switch res.Status {
				case channelcreation.StatusCreated:
					fmt.Printf("Created - channel: %s\n", res.ChannelName)
					fmt.Printf("Members: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Printf("Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusAlreadyExists:
					fmt.Printf("Already exists - channel: %s\n", res.ChannelName)
				case channelcreation.StatusMembersEnsured:
					fmt.Printf("Members ensured - channel: %s\n", res.ChannelName)
					fmt.Printf("Members ensured: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Printf("Owners ensured: %s\n", strings.Join(res.OwnerRefs, ", "))
				case channelcreation.StatusPartiallyEnsured:
					fmt.Printf("Partially ensured - channel: %s\n", res.ChannelName)
					fmt.Printf("Members ensured: %s\n", strings.Join(res.MemberRefs, ", "))
					fmt.Printf("Owners ensured: %s\n", strings.Join(res.OwnerRefs, ", "))
				default:
					fmt.Printf("Processed - channel: %s\n", res.ChannelName)
				}
			}
		}
	}

	if dryRun {
		fmt.Printf("Dry run completed - successful: %d, total: %d\n", successCount, len(results))
	} else {
		fmt.Printf("Channel creation completed - successful: %d, total: %d\n", successCount, len(results))
	}
}
