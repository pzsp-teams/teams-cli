package main

import (
	"github.com/pzsp-teams/cli/cmd"
	"github.com/pzsp-teams/cli/internal/pepper"
	"github.com/pzsp-teams/lib"
)

func main() {
	pepper.EnsurePepper()
	defer lib.Close()
	cmd.Execute()
}
