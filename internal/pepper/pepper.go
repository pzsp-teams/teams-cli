package pepper

import (
	"fmt"
	"os"

	"github.com/pzsp-teams/lib/setup"
)

// EnsurePepper checks if a pepper is set, and if not, prompts the user to set one
func EnsurePepper() {
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
