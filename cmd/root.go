package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pzsp-teams/cli/cmd/channels"
	"github.com/pzsp-teams/cli/cmd/chats"
	"github.com/pzsp-teams/cli/cmd/teams"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/logger"
)

// RootCmd is the root command for the CLI
var RootCmd = &cobra.Command{
	Use:              "teams-cli",
	Short:            "Microsoft Teams CLI tool",
	Long:             `A command-line tool for interacting with Microsoft Teams channels and chats`,
	PersistentPreRun: initializeLogger,
}

// Execute runs the root command
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v, -vv, -vvv)")

	if err := viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind verbose flag: %v\n", err)
		os.Exit(1)
	}

	RootCmd.AddCommand(teams.NewCommand())
	RootCmd.AddCommand(channels.NewCommand())
	RootCmd.AddCommand(chats.NewCommand())
}

func initializeLogger(cmd *cobra.Command, args []string) {
	verboseCount := viper.GetInt("verbose")

	logFile, err := os.OpenFile("teams-cli.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	stderrLevel := mapVerboseToLevel(verboseCount)

	initializers.InitMultiOutputLogger(initializers.MultiOutputConfig{
		StderrLevel:         stderrLevel,
		FileLevel:           logger.LevelDebug,
		FileWriter:          logFile,
		StderrOmitTimestamp: verboseCount == 0,
		FileOmitTimestamp:   false,
	})
}

func mapVerboseToLevel(count int) logger.Level {
	switch count {
	case 0:
		return logger.LevelError
	case 1:
		return logger.LevelWarn
	case 2:
		return logger.LevelInfo
	default:
		return logger.LevelDebug
	}
}
