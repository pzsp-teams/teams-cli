package main

import (
	"fmt"
	"os"

	"github.com/pzsp-teams/cli/cmd"
	"github.com/pzsp-teams/lib"
	"github.com/pzsp-teams/lib/setup"
)

func ensurePepper() {
	if !setup.PepperExists() {
		var pepper string
		print("Set a pepper for hashing passwords: ")
		_, err := fmt.Scanln(&pepper)
		if err != nil {
			fmt.Println("Error reading pepper:", err)
			os.Exit(1)
		}
		err = setup.SetPepper(pepper)
		if err != nil {
			fmt.Println("Error setting pepper:", err)
			os.Exit(1)
		}
	}
}

func main() {
	ensurePepper()
	defer lib.Close()
	cmd.Execute()
}
