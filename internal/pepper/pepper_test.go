package pepper

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsurePepper_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		exists          bool
		input           string
		setErr          error
		wantSetCalled   bool
		wantSetValue    string
		wantExitCalled  bool
		wantExitCode    int
		wantOutContains []string
	}{
		{
			name:            "pepper already exists -> no prompt, no set, no exit",
			exists:          true,
			wantSetCalled:   false,
			wantExitCalled:  false,
			wantOutContains: []string{},
		},
		{
			name:           "pepper missing -> reads input and sets",
			exists:         false,
			input:          "sekret\n",
			wantSetCalled:  true,
			wantSetValue:   "sekret",
			wantExitCalled: false,
			wantOutContains: []string{
				"Set a pepper for hashing passwords:",
			},
		},
		{
			name:           "pepper missing -> scan error -> exit 1",
			exists:         false,
			input:          "",
			wantSetCalled:  false,
			wantExitCalled: true,
			wantExitCode:   1,
			wantOutContains: []string{
				"Set a pepper for hashing passwords:",
				"Error reading pepper:",
			},
		},
		{
			name:           "pepper missing -> set error -> exit 1",
			exists:         false,
			input:          "sekret\n",
			setErr:         errors.New("set failed"),
			wantSetCalled:  true,
			wantSetValue:   "sekret",
			wantExitCalled: true,
			wantExitCode:   1,
			wantOutContains: []string{
				"Set a pepper for hashing passwords:",
				"Error setting pepper:",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			in := bytes.NewBufferString(tt.input)

			setCalled := 0
			var setValue string

			exitCalled := false
			exitCode := 0

			d := deps{
				PepperExists: func() bool { return tt.exists },
				SetPepper: func(p string) error {
					setCalled++
					setValue = p
					return tt.setErr
				},
				In:   in,
				Out:  &out,
				Exit: func(code int) { exitCalled = true; exitCode = code },
			}

			ensurePepper(d)

			require.Equal(t, tt.wantSetCalled, setCalled > 0)
			if tt.wantSetCalled {
				require.Equal(t, tt.wantSetValue, setValue)
			}

			require.Equal(t, tt.wantExitCalled, exitCalled)
			if tt.wantExitCalled {
				require.Equal(t, tt.wantExitCode, exitCode)
			}

			gotOut := out.String()
			if tt.exists {
				require.Equal(t, "", gotOut)
			} else {
				for _, sub := range tt.wantOutContains {
					require.Contains(t, gotOut, sub)
				}
			}
		})
	}
}
