package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ConfigFormat represents supported configuration file formats
type ConfigFormat string

// Accepted config file formats
const (
	FormatJSON ConfigFormat = "json"
	FormatYAML ConfigFormat = "yaml"
	FormatTOML ConfigFormat = "toml"
)

// ConfigFormatFlags represents the format flag values
type ConfigFormatFlags struct {
	JSON bool
	YAML bool
	TOML bool
}

// AddFormatFlags adds format flags to a cobra command
func AddFormatFlags(cmd *cobra.Command, flags *ConfigFormatFlags) {
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "Use JSON format")
	cmd.Flags().BoolVar(&flags.YAML, "yaml", false, "Use YAML format")
	cmd.Flags().BoolVar(&flags.TOML, "toml", false, "Use TOML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml", "toml")
}

// GetFormat returns the selected format or an error if none/multiple are selected
func (f *ConfigFormatFlags) GetFormat() (ConfigFormat, error) {
	count := 0
	var format ConfigFormat

	if f.JSON {
		count++
		format = FormatJSON
	}
	if f.YAML {
		count++
		format = FormatYAML
	}
	if f.TOML {
		count++
		format = FormatTOML
	}

	if count == 0 {
		return FormatTOML, nil
	}
	if count > 1 {
		return "", fmt.Errorf("multiple formats specified, please choose only one")
	}

	return format, nil
}
