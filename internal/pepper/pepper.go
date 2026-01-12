package pepper

import (
	"fmt"
	"io"
	"os"

	"github.com/pzsp-teams/lib/setup"
)

type deps struct {
	PepperExists func() bool
	SetPepper    func(string) error
	In           io.Reader
	Out          io.Writer
	Exit         func(int)
}

// EnsurePepper checks if a pepper is set, and if not, prompts the user to set one.
func EnsurePepper() {
	ensurePepper(deps{
		PepperExists: setup.PepperExists,
		SetPepper:    setup.SetPepper,
		In:           os.Stdin,
		Out:          os.Stdout,
		Exit:         os.Exit,
	})
}

func ensurePepper(d deps) {
	if d.PepperExists() {
		return
	}

	_, _ = fmt.Fprint(d.Out, "Set a pepper for hashing passwords: ")

	var pepper string
	_, err := fmt.Fscanln(d.In, &pepper)
	if err != nil {
		_, _ = fmt.Fprintln(d.Out, "Error reading pepper:", err)
		d.Exit(1)
		return
	}

	if err := d.SetPepper(pepper); err != nil {
		_, _ = fmt.Fprintln(d.Out, "Error setting pepper:", err)
		d.Exit(1)
		return
	}
}
