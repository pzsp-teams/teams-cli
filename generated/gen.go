package main

import (
	"log"

	"github.com/pzsp-teams/teams-cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Generate markdown docs in the current directory
	err := doc.GenMarkdownTree(cmd.RootCmd, ".")
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}
}
