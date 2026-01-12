package main

import (
	"github.com/pzsp-teams/lib"
	"github.com/pzsp-teams/teams-cli/cmd"
	"github.com/pzsp-teams/teams-cli/internal/pepper"
)

func main() {
	pepper.EnsurePepper()
	defer lib.Close()
	cmd.Execute()
}
