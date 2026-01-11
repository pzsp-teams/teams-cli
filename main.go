package main

import (
	"fmt"

	"github.com/pzsp-teams/cli/cmd"
	"github.com/pzsp-teams/lib"
	"github.com/pzsp-teams/lib/setup"
)

func init() {
	if !setup.PepperExists() {
		var pepper string
		print("Set a pepper for hashing passwords: ")
		fmt.Scanln(&pepper)
		setup.SetPepper(pepper)
	}
}

func main() {
	defer lib.Close()
	cmd.Execute()
}
