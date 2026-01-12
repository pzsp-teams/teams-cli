package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pzsp-teams/cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration files",
	Long:  `Generate and manage configuration files for the Teams CLI`,
}

var configInitCmd = &cobra.Command{
	Use:   "init [filename]",
	Short: "Generate a default configuration file",
	Long:  `Generate a default configuration file with empty values in the specified format`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigInit,
}

var (
	configFormatFlags ConfigFormatFlags
	outputPath        string
)

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)

	AddFormatFlags(configInitCmd, &configFormatFlags)

	defaultConfigPath := filepath.Join(xdg.ConfigHome, "teams-cli", "config.<format>")
	configInitCmd.Flags().StringVarP(&outputPath, "output", "o", "", fmt.Sprintf("Output file path (default: %s)", defaultConfigPath))

	RootCmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	format, err := configFormatFlags.GetFormat()
	if err != nil {
		return err
	}

	outPath := outputPath
	if outPath == "" {
		if len(args) > 0 {
			outPath = args[0]
		} else {
			configDir := filepath.Join(xdg.ConfigHome, "teams-cli")
			filename := fmt.Sprintf("config.%s", format)
			outPath = filepath.Join(configDir, filename)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content, err := config.CreateDefaultConfig(string(format))
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created default configuration file: %s\n", outPath)
	return nil
}
