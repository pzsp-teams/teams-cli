package main

import (
	"github.com/pzsp-teams/cli/cmd"
	"github.com/pzsp-teams/lib"
)

func main() {
	defer lib.Close()
	cmd.Execute()
}
