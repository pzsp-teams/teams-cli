package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pzsp-teams/teams-cli/app"
	"github.com/pzsp-teams/teams-cli/internal/client"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
	"github.com/pzsp-teams/teams-cli/internal/logger"
	"github.com/pzsp-teams/teams-cli/tui"
)

// RootCmd is the root command for the CLI
var RootCmd = &cobra.Command{
	Use:   "teams-cli",
	Short: "Microsoft Teams CLI tool",
	Long:  `A command-line tool for interacting with Microsoft Teams channels and chats`,
	Run:   runTUI,
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

	for _, def := range app.Registry {
		RootCmd.AddCommand(toCobraCommand(&def))
	}
}

func toCobraCommand(def *app.CommandDef) *cobra.Command {
	cmd := &cobra.Command{
		Use:    def.Use,
		Short:  def.Short,
		Long:   def.Long,
		PreRun: initializeLoggerAndClient,
	}

	for i := range def.Flags {
		flag := &def.Flags[i]
		switch flag.Type {
		case app.InputBool:
			cmd.Flags().BoolP(flag.Name, flag.Shorthand, false, flag.Usage)
		case app.InputInt:
			cmd.Flags().IntP(flag.Name, flag.Shorthand, 0, flag.Usage)
		case app.InputList:
			cmd.Flags().StringSliceP(flag.Name, flag.Shorthand, []string{}, flag.Usage)
		default:
			cmd.Flags().StringP(flag.Name, flag.Shorthand, "", flag.Usage)
		}

		if flag.Required {
			_ = cmd.MarkFlagRequired(flag.Name)
		}
	}

	for _, sub := range def.SubCommands {
		cmd.AddCommand(toCobraCommand(&sub))
	}

	if def.Handler != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			flags := make(map[string]any)
			for i := range def.Flags {
				f := &def.Flags[i]
				var val any
				var err error
				switch f.Type {
				case app.InputBool:
					val, err = cmd.Flags().GetBool(f.Name)
				case app.InputInt:
					val, err = cmd.Flags().GetInt(f.Name)
				case app.InputList:
					val, err = cmd.Flags().GetStringSlice(f.Name)
				default:
					val, err = cmd.Flags().GetString(f.Name)
				}
				if err != nil {
					return err
				}
				flags[f.Name] = val
			}

			_, err := def.Handler(cmd.Context(), os.Stdout, flags)
			return err
		}
	} else {
		cmd.Run = func(c *cobra.Command, args []string) {
			fullPath := c.CommandPath()
			rootName := c.Root().Name()
			startPath := fullPath
			if after, ok := strings.CutPrefix(fullPath, rootName); ok {
				startPath = after
				startPath = strings.TrimSpace(startPath)
			}
			startTUI(c.Context(), startPath)
		}
	}

	return cmd
}

func initializeLogger(cmd *cobra.Command, args []string) {
	verboseCount := viper.GetInt("verbose")

	logFile, err := os.OpenFile("teams-cli.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	stderrLevel := mapVerboseToLevel(verboseCount)

	initializers.StderrLogger = logger.NewCharmFromConfig(&logger.Config{
		Level:         stderrLevel,
		Format:        logger.FormatText,
		Output:        os.Stderr,
		OmitTimestamp: verboseCount == 0,
		AddSource:     false,
	})

	fileLogger := logger.NewCharmFromConfig(&logger.Config{
		Level:         logger.LevelDebug,
		Format:        logger.FormatText,
		Output:        logFile,
		OmitTimestamp: false,
		AddSource:     false,
	})

	initializers.Logger = logger.NewMultiLogger(initializers.StderrLogger, fileLogger)
}

func initializeClient(cmd *cobra.Command) {
	configPath := getConfigPath(cmd)
	if err := client.Initialize(cmd.Context(), configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Teams client: %v\n", err)
		os.Exit(1)
	}
}

func initializeLoggerAndClient(cmd *cobra.Command, args []string) {
	initializeLogger(cmd, args)
	initializeClient(cmd)
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

func runTUI(cmd *cobra.Command, args []string) {
	startPath := ""
	if len(args) > 0 {
		startPath = args[0]
	}
	startTUI(cmd.Context(), startPath)
}

func startTUI(ctx context.Context, startPath string) {
	initializers.DisableStderrLogger()

	if err := tui.Run(ctx, app.Registry, startPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// getConfigPath extracts the config path from the command flags
func getConfigPath(cmd *cobra.Command) string {
	configPath, _ := cmd.Flags().GetString("config")
	return configPath
}
