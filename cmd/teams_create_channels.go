package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	channelcreation "github.com/pzsp-teams/cli/internal/channel_creation"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var teamsCreateChannelsCmd = &cobra.Command{
	Use:   "create-channels",
	Short: "Create Teams channels from a data file",
	Long: `Create multiple Teams channels with members from a data file (YAML/JSON/TOML/CSV).

The data file should contain channel definitions with team_ref, channel_ref, role, and user_ref.

Examples:
  # Create channels from YAML file
  cli teams create-channels --data channels.yaml

  # Create channels from JSON file
  cli teams create-channels --data channels.json

  # Dry run to preview
  cli teams create-channels --data channels.yaml --dry-run`,
	RunE: runTeamsCreateChannels,
}

var (
	teamRef              string
	createChannelsData   string
	createChannelsDryRun bool
)

func init() {
	teamsCreateChannelsCmd.Flags().StringVar(&teamRef, "team", "", "Name of the team in which to create channels")
	teamsCreateChannelsCmd.Flags().StringVar(&createChannelsData, "data", "", "Path to channels data file (YAML/JSON/TOML/CSV)")
	teamsCreateChannelsCmd.Flags().BoolVar(&createChannelsDryRun, "dry-run", false, "Preview without creating channels")

	if err := teamsCreateChannelsCmd.MarkFlagRequired("data"); err != nil {
		panic(fmt.Sprintf("failed to mark data flag as required: %v", err))
	}
	if err := teamsCreateChannelsCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runTeamsCreateChannels(cmd *cobra.Command, args []string) error {
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
	results := teamsClient.ChannelCreator.CreateChannels(ctx, teamRef, channelData, true, createChannelsDryRun)

	successCount := 0
	for _, res := range results {
		if res.Error != nil {
			log.Error("Failed", "channel", res.ChannelName, "error", res.Error)
		} else {
			successCount++
			if createChannelsDryRun {
				log.Info("Would create", "channel", res.ChannelName)
			} else {
				log.Info("Created", "channel", res.ChannelName, "id", res.ChannelID)
			}
		}
	}

	if createChannelsDryRun {
		log.Info("Dry run completed", "successful", successCount, "total", len(results))
	} else {
		log.Info("Channel creation completed", "successful", successCount, "total", len(results))
	}

	return nil
}
