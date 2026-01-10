package channels

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
	channelcreation "github.com/pzsp-teams/cli/internal/channels/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
)

type createFlags struct {
	team             string
	data             string
	dryRun           bool
	ensureInChannels bool
	ensureInTeam     bool
}

func newCreateCommand() *cobra.Command {
	flags := &createFlags{}

	cmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "Name of the team in which to create channels")
	cmd.Flags().StringVar(&flags.data, "data", "", "Path to channels data file (YAML/JSON/TOML/CSV)")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview without creating channels")
	cmd.Flags().BoolVar(&flags.ensureInChannels, "ensure-in-channels", false, "Ensure members are added to channels if they already exist")
	cmd.Flags().BoolVar(&flags.ensureInTeam, "ensure-in-team", false, "Ensure members are members of the team")

	if err := cmd.MarkFlagRequired("data"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark data flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runCreate(cmd *cobra.Command, flags *createFlags) error {
	log := initializers.Logger
	ctx := cmd.Context()

	dataFile, err := os.Open(flags.data)
	if err != nil {
		log.Error("Failed to open data file", "file", flags.data, "error", err)
		return fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(flags.data), ".")

	log.Debug("Parsing channels data", "file", flags.data)
	channelData, err := channelcreation.ParseChannelsDataByExtension(dataFile, extension)
	if err != nil {
		log.Error("Failed to parse channels data", "error", err)
		return fmt.Errorf("failed to parse channels data: %w", err)
	}
	_ = dataFile.Close()

	log.Info("Parsed channels data", "channels", len(channelData))

	teamsClient, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	log.Info("Creating channels", "count", len(channelData), "dryRun", flags.dryRun)
	results := teamsClient.Channels.Create(ctx, flags.team, channelData, flags.ensureInChannels, flags.ensureInTeam, flags.dryRun)

	printChannelCreationResults(results, flags.dryRun)

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
